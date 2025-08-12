package syncproviders

import (
	"context"
	"fmt"
	"log/slog"

	"github.com/urfave/cli/v3"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"

	"github.com/opentofu/registry-ui/pkg/config"
	"github.com/opentofu/registry-ui/pkg/docscraper"
	"github.com/opentofu/registry-ui/pkg/index"
)

func NewCommand(cfg *config.BackendConfig) *cli.Command {
	return &cli.Command{
		Name:  "sync-providers",
		Usage: "Sync multiple providers matching a filter pattern from the registry",
		Flags: []cli.Flag{
			&cli.StringFlag{
				Name:     "filter",
				Aliases:  []string{"f"},
				Usage:    "Provider filter pattern (e.g., 'hashicorp/*', '*/aws', 'hashicorp/aws'). Supports wildcards.",
				Required: true,
			},
			&cli.StringFlag{
				Name:     "version",
				Aliases:  []string{"v"},
				Usage:    "Specific version to sync across all matching providers (e.g., 6.7.0). If not provided, syncs all missing versions",
				Required: false,
			},
		},
		Action: func(ctx context.Context, cmd *cli.Command) error {
			return run(ctx, cmd, cfg)
		},
	}
}

func run(ctx context.Context, cmd *cli.Command, cfg *config.BackendConfig) error {
	tracer := otel.Tracer("opentofu-registry-backend")
	ctx, span := tracer.Start(ctx, "sync-providers")
	defer span.End()

	filter := cmd.String("filter")
	specificVersion := cmd.String("version")

	span.SetAttributes(
		attribute.String("filter", filter),
	)

	if specificVersion != "" {
		span.SetAttributes(attribute.String("version", specificVersion))
		slog.InfoContext(ctx, "Syncing providers with specific version from registry",
			"filter", filter, "version", specificVersion)
	} else {
		slog.InfoContext(ctx, "Syncing providers from registry",
			"filter", filter)
	}

	// Set up dependencies
	pool, err := cfg.DB.GetPool(ctx)
	if err != nil {
		span.RecordError(err)
		slog.ErrorContext(ctx, "Failed to get database pool", "error", err)
		return fmt.Errorf("failed to connect to database: %w", err)
	}

	s3Client, err := cfg.Bucket.GetClient(ctx)
	if err != nil {
		span.RecordError(err)
		slog.ErrorContext(ctx, "Failed to get S3 client", "error", err)
		return fmt.Errorf("failed to connect to s3 client: %w", err)
	}

	scraper := docscraper.New(cfg, s3Client, pool)
	indexService := index.NewIndexService(cfg, pool, scraper, s3Client)

	_, err = indexService.IndexMultipleProviders(ctx, filter, specificVersion)
	if err != nil {
		span.RecordError(err)
		return err
	}

	return nil
}
