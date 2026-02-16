// Package syncprovider implements the command to sync all versions of a specific provider from the registry.
// It will fetch missing versions, update metadata, and handle documentation scraping.
package syncprovider

import (
	"context"
	"fmt"
	"log/slog"

	"github.com/urfave/cli/v3"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/codes"

	"github.com/opentofu/registry-ui/pkg/config"
	"github.com/opentofu/registry-ui/pkg/provider"
	"github.com/opentofu/registry-ui/pkg/registry"
	"github.com/opentofu/registry-ui/pkg/telemetry"
)

func NewCommand() *cli.Command {
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
			return run(ctx, cmd)
		},
	}
}

func run(ctx context.Context, cmd *cli.Command) error {
	cfg := config.FromCLI(cmd)
	ctx, span := telemetry.Tracer().Start(ctx, "sync-provider")
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

	// Fetch provider data from registry
	registryClient, err := registry.New(cfg.RegistryPath)
	if err != nil {
		span.RecordError(err)
		span.SetStatus(codes.Error, err.Error())
		return fmt.Errorf("failed to create registry client: %w", err)
	}

	err = registryClient.Update(ctx)
	if err != nil {
		span.RecordError(err)
		span.SetStatus(codes.Error, err.Error())
		return fmt.Errorf("failed to update registry data: %w", err)
	}

	prov, err := registryClient.GetProvider(ctx, namespace, name)
	if err != nil {
		span.RecordError(err)
		span.SetStatus(codes.Error, err.Error())
		return fmt.Errorf("provider %s/%s not found in registry: %w", namespace, name, err)
	}

	// Create provider reader
	providerReader, err := provider.NewProviderReader(cfg)
	if err != nil {
		span.RecordError(err)
		span.SetStatus(codes.Error, err.Error())
		slog.ErrorContext(ctx, "Failed to create provider reader", "error", err)
		return fmt.Errorf("failed to create provider reader: %w", err)
	}

	// Scrape the provider version(s)
	if specificVersion != "" {
		_, err = providerReader.IndexVersion(ctx, prov, specificVersion)
	} else {
		err = providerReader.ScrapeAllVersions(ctx, prov)
	}

	if err != nil {
		span.RecordError(err)
		span.SetStatus(codes.Error, err.Error())
		return err
	}

	return nil
}
