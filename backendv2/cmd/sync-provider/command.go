package syncprovider

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
		Name:  "sync-provider",
		Usage: "Sync all versions for a specific provider from the registry",
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
				Usage:    "Specific version to sync (e.g., 6.7.0). If not provided, syncs all missing versions",
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
	ctx, span := tracer.Start(ctx, "sync-provider")
	defer span.End()

	namespace := cmd.String("namespace")
	name := cmd.String("name")
	specificVersion := cmd.String("version")

	span.SetAttributes(
		attribute.String("provider.namespace", namespace),
		attribute.String("provider.name", name),
	)

	if specificVersion != "" {
		span.SetAttributes(attribute.String("provider.version", specificVersion))
		slog.InfoContext(ctx, "Syncing specific provider version from registry",
			"namespace", namespace, "name", name, "version", specificVersion)
	} else {
		slog.InfoContext(ctx, "Syncing provider versions from registry",
			"namespace", namespace, "name", name)
	}

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

	_, err = indexService.IndexProviderVersion(ctx, namespace, name, specificVersion)
	if err != nil {
		span.RecordError(err)
		return err
	}

	return nil
}
