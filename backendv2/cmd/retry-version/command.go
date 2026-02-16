// Package retryversion implements the command to reset a skipped or failed version for retry
package retryversion

import (
	"context"
	"fmt"
	"log/slog"

	"github.com/urfave/cli/v3"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/codes"

	"github.com/opentofu/registry-ui/pkg/config"
	"github.com/opentofu/registry-ui/pkg/telemetry"
)

func NewCommand() *cli.Command {
	return &cli.Command{
		Name:  "retry-version",
		Usage: "Reset a skipped or failed version to allow retry during next sync",
		Description: `This command resets the status of a version from 'skipped' or 'failed' back to 'pending'.
This will cause the version to be processed again during the next sync operation.

Note: This command deletes the version record from the database, which will cause
it to be re-processed from scratch during the next sync.`,
		Flags: []cli.Flag{
			&cli.StringFlag{
				Name:     "type",
				Aliases:  []string{"t"},
				Usage:    "Resource type: 'provider' or 'module'",
				Required: true,
			},
			&cli.StringFlag{
				Name:     "namespace",
				Aliases:  []string{"n"},
				Usage:    "Namespace (e.g., hashicorp)",
				Required: true,
			},
			&cli.StringFlag{
				Name:     "name",
				Usage:    "Name (e.g., aws for provider, vpc for module)",
				Required: true,
			},
			&cli.StringFlag{
				Name:     "target",
				Usage:    "Target provider/platform from the module address namespace/name/target (e.g., 'aws' in hashicorp/vpc/aws) - required for modules only",
				Required: false,
			},
			&cli.StringFlag{
				Name:     "version",
				Aliases:  []string{"v"},
				Usage:    "Version to retry (e.g., 1.2.3)",
				Required: true,
			},
		},
		Action: func(ctx context.Context, cmd *cli.Command) error {
			return run(ctx, cmd)
		},
	}
}

func run(ctx context.Context, cmd *cli.Command) error {
	cfg := config.FromCLI(cmd)
	ctx, span := telemetry.Tracer().Start(ctx, "retry-version")
	defer span.End()

	resourceType := cmd.String("type")
	namespace := cmd.String("namespace")
	name := cmd.String("name")
	target := cmd.String("target")
	version := cmd.String("version")

	// Validate resource type
	if resourceType != "provider" && resourceType != "module" {
		err := fmt.Errorf("invalid type: %s (must be 'provider' or 'module')", resourceType)
		span.RecordError(err)
		span.SetStatus(codes.Error, err.Error())
		return err
	}

	// Validate target for modules
	if resourceType == "module" && target == "" {
		err := fmt.Errorf("--target is required for modules")
		span.RecordError(err)
		span.SetStatus(codes.Error, err.Error())
		return err
	}

	span.SetAttributes(
		attribute.String("resource.type", resourceType),
		attribute.String("resource.namespace", namespace),
		attribute.String("resource.name", name),
		attribute.String("resource.version", version),
	)

	if resourceType == "module" {
		span.SetAttributes(attribute.String("resource.target", target))
	}

	// Connect to database
	pool, err := cfg.DB.GetPool(ctx)
	if err != nil {
		span.RecordError(err)
		span.SetStatus(codes.Error, err.Error())
		slog.ErrorContext(ctx, "Failed to connect to database", "error", err)
		return fmt.Errorf("failed to connect to database: %w", err)
	}

	// Delete the version record to allow retry
	// The scraper will recreate it during the next sync
	var deleteQuery string
	var args []interface{}

	if resourceType == "provider" {
		slog.InfoContext(ctx, "Resetting provider version for retry",
			"namespace", namespace, "name", name, "version", version)

		deleteQuery = `DELETE FROM provider_versions WHERE provider_namespace = $1 AND provider_name = $2 AND version = $3`
		args = []interface{}{namespace, name, version}

		result, err := pool.Exec(ctx, deleteQuery, args...)
		if err != nil {
			span.RecordError(err)
			span.SetStatus(codes.Error, err.Error())
			slog.ErrorContext(ctx, "Failed to delete provider version", "error", err)
			return fmt.Errorf("failed to delete provider version: %w", err)
		}

		if result.RowsAffected() == 0 {
			err := fmt.Errorf("provider version %s/%s@%s not found", namespace, name, version)
			span.RecordError(err)
			span.SetStatus(codes.Error, err.Error())
			return err
		}

		slog.InfoContext(ctx, "Successfully reset provider version for retry",
			"namespace", namespace, "name", name, "version", version)
		fmt.Printf("✓ Reset provider %s/%s@%s - it will be retried during next sync\n", namespace, name, version)
	} else {
		slog.InfoContext(ctx, "Resetting module version for retry",
			"namespace", namespace, "name", name, "target", target, "version", version)

		deleteQuery = `DELETE FROM module_versions WHERE module_namespace = $1 AND module_name = $2 AND module_target = $3 AND version = $4`
		args = []interface{}{namespace, name, target, version}

		result, err := pool.Exec(ctx, deleteQuery, args...)
		if err != nil {
			span.RecordError(err)
			span.SetStatus(codes.Error, err.Error())
			slog.ErrorContext(ctx, "Failed to delete module version", "error", err)
			return fmt.Errorf("failed to delete module version: %w", err)
		}

		if result.RowsAffected() == 0 {
			err := fmt.Errorf("module version %s/%s/%s@%s not found", namespace, name, target, version)
			span.RecordError(err)
			span.SetStatus(codes.Error, err.Error())
			return err
		}

		slog.InfoContext(ctx, "Successfully reset module version for retry",
			"namespace", namespace, "name", name, "target", target, "version", version)
		fmt.Printf("✓ Reset module %s/%s/%s@%s - it will be retried during next sync\n", namespace, name, target, version)
	}

	return nil
}
