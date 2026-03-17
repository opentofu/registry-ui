// Package syncmodule implements the sync-module command to sync all versions for a specific module from the registry.
package syncmodule

import (
	"context"
	"fmt"
	"log/slog"

	"github.com/urfave/cli/v3"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/codes"

	"github.com/opentofu/registry-ui/pkg/config"
	"github.com/opentofu/registry-ui/pkg/module"
	"github.com/opentofu/registry-ui/pkg/registry"
	"github.com/opentofu/registry-ui/pkg/telemetry"
)

func NewCommand() *cli.Command {
	return &cli.Command{
		Name:  "sync-module",
		Usage: "Sync all versions for a specific module from the registry",
		Flags: []cli.Flag{
			&cli.StringFlag{
				Name:     "namespace",
				Usage:    "Module namespace (e.g., hashicorp)",
				Required: true,
			},
			&cli.StringFlag{
				Name:     "name",
				Usage:    "Module name (e.g., vpc)",
				Required: true,
			},
			&cli.StringFlag{
				Name:     "target",
				Usage:    "Target provider/platform from the module address namespace/name/target (e.g., 'aws' in hashicorp/vpc/aws)",
				Required: true,
			},
			&cli.StringFlag{
				Name:     "version",
				Aliases:  []string{"v"},
				Usage:    "Specific version to sync (e.g., 3.0.0). If not provided, syncs all missing versions",
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
	ctx, span := telemetry.Tracer().Start(ctx, "cmd.sync_module")
	defer span.End()

	namespace := cmd.String("namespace")
	name := cmd.String("name")
	target := cmd.String("target")
	specificVersion := cmd.String("version")

	span.SetAttributes(
		attribute.String("module.namespace", namespace),
		attribute.String("module.name", name),
		attribute.String("module.target", target),
	)

	if specificVersion != "" {
		span.SetAttributes(attribute.String("module.version", specificVersion))
		slog.InfoContext(ctx, "Syncing specific module version from registry",
			"namespace", namespace, "name", name, "target", target, "version", specificVersion)
	} else {
		slog.InfoContext(ctx, "Syncing module versions from registry",
			"namespace", namespace, "name", name, "target", target)
	}

	pool, err := cfg.DB.GetPool(ctx)
	if err != nil {
		span.RecordError(err)
		span.SetStatus(codes.Error, err.Error())
		return fmt.Errorf("failed to initialize database pool: %w", err)
	}
	defer pool.Close()

	moduleReader, err := module.NewModuleReader(ctx, cfg, pool)
	if err != nil {
		span.RecordError(err)
		span.SetStatus(codes.Error, err.Error())
		slog.ErrorContext(ctx, "Failed to create module reader", "error", err)
		return fmt.Errorf("failed to create module reader: %w", err)
	}

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
		return fmt.Errorf("failed to update registry client: %w", err)
	}

	module, err := registryClient.GetModule(ctx, namespace, name, target)
	if err != nil {
		span.RecordError(err)
		span.SetStatus(codes.Error, err.Error())
		return fmt.Errorf("module %s/%s/%s not found in registry: %w", namespace, name, target, err)
	}

	if specificVersion != "" {
		err = moduleReader.ScrapeVersion(ctx, module, specificVersion)
		if err != nil {
			span.RecordError(err)
			span.SetStatus(codes.Error, err.Error())
			slog.ErrorContext(ctx, "Failed to scrape module version", "namespace", namespace, "name", name, "target", target, "version", specificVersion, "error", err)
			return fmt.Errorf("failed to scrape version %s for module %s/%s/%s: %w", specificVersion, namespace, name, target, err)
		}
		return nil
	}

	err = moduleReader.ScrapeAllVersions(ctx, module)
	if err != nil {
		span.RecordError(err)
		slog.ErrorContext(ctx, "Failed to scrape all module versions", "namespace", namespace, "name", name, "target", target, "error", err)
		return fmt.Errorf("failed to scrape all versions for module %s/%s/%s: %w",
			namespace, name, target, err)
	}

	return nil
}
