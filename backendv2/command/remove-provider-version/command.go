// Package removeproviderversion implements the command to remove a provider version from the database and S3
package removeproviderversion

import (
	"context"
	"fmt"
	"log/slog"

	"github.com/jackc/pgx/v5"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/s3"
	"github.com/aws/aws-sdk-go-v2/service/s3/types"
	"github.com/urfave/cli/v3"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/codes"

	"github.com/opentofu/registry-ui/pkg/config"
	"github.com/opentofu/registry-ui/pkg/telemetry"
)

func NewCommand() *cli.Command {
	return &cli.Command{
		Name:  "remove-provider-version",
		Usage: "Remove a provider version from the database and S3",
		Flags: []cli.Flag{
			&cli.StringFlag{
				Name:     "namespace",
				Aliases:  []string{"n"},
				Usage:    "Provider namespace (e.g., hashicorp)",
				Required: true,
			},
			&cli.StringFlag{
				Name:     "name",
				Usage:    "Provider name (e.g., aws)",
				Required: true,
			},
			&cli.StringFlag{
				Name:     "version",
				Aliases:  []string{"v"},
				Usage:    "Version to remove (e.g., 1.2.3)",
				Required: true,
			},
			&cli.BoolFlag{
				Name:  "dry-run",
				Usage: "Preview what will be deleted without actually deleting",
			},
		},
		Action: func(ctx context.Context, cmd *cli.Command) error {
			return run(ctx, cmd)
		},
	}
}

func run(ctx context.Context, cmd *cli.Command) error {
	cfg := config.FromCLI(cmd)
	ctx, span := telemetry.Tracer().Start(ctx, "cmd.remove_provider_version")
	defer span.End()

	namespace := cmd.String("namespace")
	name := cmd.String("name")
	version := cmd.String("version")
	dryRun := cmd.Bool("dry-run")

	span.SetAttributes(
		attribute.String("provider.namespace", namespace),
		attribute.String("provider.name", name),
		attribute.String("provider.version", version),
		attribute.Bool("dry_run", dryRun),
	)

	slog.InfoContext(ctx, "Removing provider version",
		"namespace", namespace, "name", name, "version", version, "dry_run", dryRun)

	// Connect to database
	pool, err := cfg.DB.GetPool(ctx)
	if err != nil {
		span.RecordError(err)
		span.SetStatus(codes.Error, err.Error())
		slog.ErrorContext(ctx, "Failed to connect to database", "error", err)
		return fmt.Errorf("failed to connect to database: %w", err)
	}
	defer pool.Close()

	// Get S3 client
	s3Client, err := cfg.Bucket.GetClient(ctx)
	if err != nil {
		span.RecordError(err)
		span.SetStatus(codes.Error, err.Error())
		return fmt.Errorf("failed to get S3 client: %w", err)
	}

	// Start transaction for consistent reads + delete
	tx, err := pool.Begin(ctx)
	if err != nil {
		span.RecordError(err)
		span.SetStatus(codes.Error, err.Error())
		slog.ErrorContext(ctx, "Failed to start transaction", "error", err)
		return fmt.Errorf("failed to start transaction: %w", err)
	}
	defer tx.Rollback(ctx) //nolint:errcheck

	// Check if version exists and count related records
	docCount, licenseCount, err := queryVersionInfo(ctx, tx, namespace, name, version)
	if err != nil {
		span.RecordError(err)
		span.SetStatus(codes.Error, err.Error())
		return err
	}

	// List S3 objects
	s3Prefix := fmt.Sprintf("providers/%s/%s/%s/", namespace, name, version)
	s3Objects, err := listS3Objects(ctx, s3Client, cfg.Bucket.BucketName, s3Prefix)
	if err != nil {
		span.RecordError(err)
		span.SetStatus(codes.Error, err.Error())
		return fmt.Errorf("failed to list S3 objects: %w", err)
	}

	printSummary(namespace, name, version, docCount, licenseCount, s3Objects, s3Prefix)

	if dryRun {
		fmt.Printf("\n[DRY RUN] No changes made.\n")
		return nil
	}

	// Delete from database (cascades to related tables)
	err = deleteVersionFromDB(ctx, tx, namespace, name, version, docCount, licenseCount)
	if err != nil {
		span.RecordError(err)
		span.SetStatus(codes.Error, err.Error())
		return err
	}

	// Delete from S3
	if len(s3Objects) > 0 {
		slog.InfoContext(ctx, "Deleting from S3", "prefix", s3Prefix, "count", len(s3Objects))
		deleted, err := deleteS3Objects(ctx, s3Client, cfg.Bucket.BucketName, s3Objects)
		if err != nil {
			span.RecordError(err)
			span.SetStatus(codes.Error, err.Error())
			return fmt.Errorf("failed to delete S3 objects (DB already deleted): %w", err)
		}
		slog.InfoContext(ctx, "Deleted from S3", "count", deleted)
	}

	fmt.Printf("\n✓ Successfully removed provider version %s/%s@%s\n", namespace, name, version)
	return nil
}

func queryVersionInfo(ctx context.Context, tx pgx.Tx, namespace, name, version string) (docCount int, licenseCount int, err error) {
	var exists bool
	err = tx.QueryRow(ctx, `
		SELECT
			EXISTS(SELECT 1 FROM provider_versions
				WHERE provider_namespace = $1 AND provider_name = $2 AND version = $3),
			(SELECT COUNT(*) FROM provider_documents
				WHERE provider_namespace = $1 AND provider_name = $2 AND version = $3),
			(SELECT COUNT(*) FROM provider_version_licenses
				WHERE provider_namespace = $1 AND provider_name = $2 AND version = $3)`,
		namespace, name, version).Scan(&exists, &docCount, &licenseCount)
	if err != nil {
		return 0, 0, fmt.Errorf("failed to query version info: %w", err)
	}

	if !exists {
		return 0, 0, fmt.Errorf("provider version %s/%s@%s not found in database", namespace, name, version)
	}

	return docCount, licenseCount, nil
}

func deleteVersionFromDB(ctx context.Context, tx pgx.Tx, namespace, name, version string, docCount, licenseCount int) error {
	slog.InfoContext(ctx, "Deleting from database", "namespace", namespace, "name", name, "version", version)

	result, err := tx.Exec(ctx, `
		DELETE FROM provider_versions
		WHERE provider_namespace = $1 AND provider_name = $2 AND version = $3`,
		namespace, name, version)
	if err != nil {
		return fmt.Errorf("failed to delete from database: %w", err)
	}

	if result.RowsAffected() == 0 {
		return fmt.Errorf("no rows deleted from database")
	}

	err = tx.Commit(ctx)
	if err != nil {
		return fmt.Errorf("failed to commit transaction: %w", err)
	}

	slog.InfoContext(ctx, "Deleted from database",
		"rows_affected", result.RowsAffected(),
		"cascaded_docs", docCount,
		"cascaded_licenses", licenseCount)

	return nil
}

func printSummary(namespace string, name string, version string, docCount int, licenseCount int, s3Objects []string, s3Prefix string) {
	fmt.Printf("\nProvider version: %s/%s@%s\n", namespace, name, version)
	fmt.Printf("Database records to delete:\n")
	fmt.Printf("  - 1 provider_versions record\n")
	fmt.Printf("  - %d provider_documents records\n", docCount)
	fmt.Printf("  - %d provider_version_licenses records\n", licenseCount)
	fmt.Printf("S3 objects to delete: %d (prefix: %s)\n", len(s3Objects), s3Prefix)
}

func listS3Objects(ctx context.Context, client *s3.Client, bucket, prefix string) ([]string, error) {
	var objects []string

	paginator := s3.NewListObjectsV2Paginator(client, &s3.ListObjectsV2Input{
		Bucket: aws.String(bucket),
		Prefix: aws.String(prefix),
	})

	for paginator.HasMorePages() {
		page, err := paginator.NextPage(ctx)
		if err != nil {
			return nil, err
		}

		for _, obj := range page.Contents {
			objects = append(objects, *obj.Key)
		}
	}

	return objects, nil
}

func deleteS3Objects(ctx context.Context, client *s3.Client, bucket string, keys []string) (int, error) {
	if len(keys) == 0 {
		return 0, nil
	}

	var totalDeleted int

	// Delete in batches of 1000 (S3 API limit)
	for i := 0; i < len(keys); i += 1000 {
		end := i + 1000
		if end > len(keys) {
			end = len(keys)
		}

		batch := keys[i:end]
		objects := make([]types.ObjectIdentifier, len(batch))
		for j, key := range batch {
			objects[j] = types.ObjectIdentifier{Key: aws.String(key)}
		}

		_, err := client.DeleteObjects(ctx, &s3.DeleteObjectsInput{
			Bucket: aws.String(bucket),
			Delete: &types.Delete{Objects: objects},
		})
		if err != nil {
			return totalDeleted, err
		}

		totalDeleted += len(batch)
	}

	return totalDeleted, nil
}
