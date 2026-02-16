// Package rebuildglobalindexes implements the command to rebuild global provider and module indexes from the database
package rebuildglobalindexes

import (
	"context"
	"fmt"
	"log/slog"

	"github.com/urfave/cli/v3"
	"go.opentelemetry.io/otel/attribute"

	"github.com/aws/aws-sdk-go-v2/feature/s3/manager"
	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/opentofu/registry-ui/pkg/config"
	"github.com/opentofu/registry-ui/pkg/index"
	"github.com/opentofu/registry-ui/pkg/telemetry"
)

func NewCommand() *cli.Command {
	return &cli.Command{
		Name:  "rebuild-global-indexes",
		Usage: "Rebuild global provider and module indexes from the database",
		Flags: []cli.Flag{
			&cli.BoolFlag{
				Name:    "providers",
				Aliases: []string{"p"},
				Usage:   "Rebuild providers global index",
				Value:   false,
			},
			&cli.BoolFlag{
				Name:    "modules",
				Aliases: []string{"m"},
				Usage:   "Rebuild modules global index",
				Value:   false,
			},
		},
		Action: func(ctx context.Context, cmd *cli.Command) error {
			return run(ctx, cmd)
		},
	}
}

func run(ctx context.Context, cmd *cli.Command) error {
	cfg := config.FromCLI(cmd)
	ctx, span := telemetry.Tracer().Start(ctx, "rebuild-global-indexes")
	defer span.End()

	rebuildProviders := cmd.Bool("providers")
	rebuildModules := cmd.Bool("modules")

	// If neither flag is specified, rebuild both
	if !rebuildProviders && !rebuildModules {
		rebuildProviders = true
		rebuildModules = true
	}

	slog.InfoContext(ctx, "Starting global index rebuild",
		"providers", rebuildProviders,
		"modules", rebuildModules)

	// Connect to database
	pool, err := cfg.DB.GetPool(ctx)
	if err != nil {
		span.RecordError(err)
		return fmt.Errorf("failed to connect to database: %w", err)
	}

	// Create S3 client for uploads
	s3Client, err := cfg.Bucket.GetClient(ctx)
	if err != nil {
		span.RecordError(err)
		return fmt.Errorf("failed to create S3 client: %w", err)
	}

	// Create uploader
	uploader := manager.NewUploader(s3Client)

	// Rebuild provider index if requested
	if rebuildProviders {
		if err := rebuildProviderIndex(ctx, pool, uploader, cfg.Bucket.BucketName); err != nil {
			span.RecordError(err)
			return fmt.Errorf("failed to rebuild provider index: %w", err)
		}
	}

	// Rebuild module index if requested
	if rebuildModules {
		if err := rebuildModuleIndex(ctx, pool, uploader, cfg.Bucket.BucketName); err != nil {
			span.RecordError(err)
			return fmt.Errorf("failed to rebuild module index: %w", err)
		}
	}

	slog.InfoContext(ctx, "Successfully rebuilt global indexes")
	return nil
}

func rebuildProviderIndex(ctx context.Context, pool *pgxpool.Pool, uploader *manager.Uploader, bucketName string) error {
	ctx, span := telemetry.Tracer().Start(ctx, "rebuild-global-indexes.providers")
	defer span.End()

	slog.InfoContext(ctx, "Rebuilding global provider index from database")

	// Query database and build global index
	globalIndex, err := index.RebuildGlobalProviderIndex(ctx, pool)
	if err != nil {
		span.RecordError(err)
		return fmt.Errorf("failed to rebuild provider index: %w", err)
	}

	span.SetAttributes(attribute.Int("providers.count", len(globalIndex.Providers)))
	slog.InfoContext(ctx, "Built global provider index from database",
		"provider_count", len(globalIndex.Providers))

	// Upload to S3
	key := "providers/index.json"
	if err := index.UploadGlobalProviderIndex(ctx, uploader, bucketName, key, globalIndex); err != nil {
		span.RecordError(err)
		return fmt.Errorf("failed to upload provider index to S3: %w", err)
	}

	slog.InfoContext(ctx, "Successfully uploaded global provider index to S3",
		"key", key,
		"provider_count", len(globalIndex.Providers))

	return nil
}

func rebuildModuleIndex(ctx context.Context, pool *pgxpool.Pool, uploader *manager.Uploader, bucketName string) error {
	ctx, span := telemetry.Tracer().Start(ctx, "rebuild-global-indexes.modules")
	defer span.End()

	slog.InfoContext(ctx, "Rebuilding global module index from database")

	// Query database and build global index
	globalIndex, err := index.RebuildGlobalModuleIndex(ctx, pool)
	if err != nil {
		span.RecordError(err)
		return fmt.Errorf("failed to rebuild module index: %w", err)
	}

	span.SetAttributes(attribute.Int("modules.count", len(globalIndex.Modules)))
	slog.InfoContext(ctx, "Built global module index from database",
		"module_count", len(globalIndex.Modules))

	// Upload to S3
	key := "modules/index.json"
	if err := index.UploadGlobalModuleIndex(ctx, uploader, bucketName, key, globalIndex); err != nil {
		span.RecordError(err)
		return fmt.Errorf("failed to upload module index to S3: %w", err)
	}

	slog.InfoContext(ctx, "Successfully uploaded global module index to S3",
		"key", key,
		"module_count", len(globalIndex.Modules))

	return nil
}
