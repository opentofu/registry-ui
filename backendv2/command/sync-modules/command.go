// Package syncmodules manages the sync-modules CLI command
package syncmodules

import (
	"context"
	"fmt"
	"log/slog"
	"sync"

	"github.com/google/uuid"
	"github.com/urfave/cli/v3"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/codes"
	"go.opentelemetry.io/otel/trace"
	"golang.org/x/sync/errgroup"

	"github.com/opentofu/registry-ui/pkg/config"
	"github.com/opentofu/registry-ui/pkg/module"
	"github.com/opentofu/registry-ui/pkg/registry"
	"github.com/opentofu/registry-ui/pkg/telemetry"
)

func NewCommand() *cli.Command {
	return &cli.Command{
		Name:  "sync-modules",
		Usage: "Sync multiple modules matching a filter pattern from the registry",
		Flags: []cli.Flag{
			&cli.StringFlag{
				Name:     "filter",
				Aliases:  []string{"f"},
				Usage:    "Module filter pattern (e.g., 'hashicorp/*', '*/vpc', 'hashicorp/vpc/aws'). Supports wildcards.",
				Required: false,
			},
			&cli.StringFlag{
				Name:     "version",
				Aliases:  []string{"v"},
				Usage:    "Specific version to sync across all matching modules (e.g., 3.0.0). If not provided, syncs all missing versions",
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
	ctx, span := telemetry.Tracer().Start(ctx, "cmd.sync_modules")
	defer span.End()

	// Generate batch job ID for correlating all module traces
	batchJobID := uuid.New().String()

	filter := cmd.String("filter")
	specificVersion := cmd.String("version")

	span.SetAttributes(
		attribute.String("filter", filter),
		attribute.String(telemetry.BatchJobIDKey, batchJobID),
		attribute.String(telemetry.BatchJobNameKey, "sync-modules"),
	)

	if specificVersion != "" {
		span.SetAttributes(attribute.String("version", specificVersion))
		slog.InfoContext(ctx, "Syncing modules with specific version from registry",
			"filter", filter, "version", specificVersion)
	} else {
		slog.InfoContext(ctx, "Syncing modules from registry",
			"filter", filter)
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
		return fmt.Errorf("failed to update registry: %w", err)
	}

	modules, err := registryClient.ListModules(ctx, filter)
	if err != nil {
		span.RecordError(err)
		span.SetStatus(codes.Error, err.Error())
		return fmt.Errorf("failed to list modules with filter %s: %w", filter, err)
	}

	if len(modules) == 0 {
		slog.InfoContext(ctx, "No modules found matching filter", "filter", filter)
		return nil
	}

	slog.InfoContext(ctx, "Found modules to sync", "count", len(modules), "filter", filter)

	// Sync each module in parallel using errgroup
	g, gctx := errgroup.WithContext(ctx)
	g.SetLimit(cfg.Concurrency.Module)

	slog.InfoContext(ctx, "Starting parallel module sync",
		"modules", len(modules),
		"concurrency", cfg.Concurrency.Module)

	var (
		successCount int
		failCount    int
		mu           sync.Mutex
	)

	for _, mod := range modules {
		g.Go(func() error {
			// Start a new trace for each module, linked back to the batch job span
			modCtx, modSpan := telemetry.LinkedSpanStart(gctx, fmt.Sprintf("module.sync.%s/%s/%s", mod.Namespace, mod.Name, mod.Target),
				trace.WithAttributes(
					attribute.String(telemetry.BatchJobIDKey, batchJobID),
					attribute.String(telemetry.BatchJobNameKey, "sync-modules"),
					attribute.String("module.namespace", mod.Namespace),
					attribute.String("module.name", mod.Name),
					attribute.String("module.target", mod.Target),
				),
			)
			defer modSpan.End()

			logger := slog.With("module.namespace", mod.Namespace, "module.name", mod.Name, "module.target", mod.Target)
			logger.InfoContext(modCtx, "Module sync started")

			// Sync the module version(s)
			var scrapeErr error
			if specificVersion != "" {
				scrapeErr = moduleReader.ScrapeVersion(modCtx, &mod, specificVersion)
			} else {
				scrapeErr = moduleReader.ScrapeAllVersions(modCtx, &mod)
			}

			mu.Lock()
			defer mu.Unlock()

			if scrapeErr != nil {
				logger.ErrorContext(modCtx, "Module sync failed", "error", scrapeErr)
				modSpan.RecordError(scrapeErr)
				failCount++
			} else {
				successCount++
				logger.InfoContext(modCtx, "Module sync completed")
			}

			return nil // Don't fail the entire operation if one module fails
		})
	}

	// Wait for all modules to complete
	if err := g.Wait(); err != nil {
		span.RecordError(err)
		span.SetStatus(codes.Error, err.Error())
		return fmt.Errorf("module sync failed: %w", err)
	}

	span.SetAttributes(
		attribute.Int("modules.total", len(modules)),
		attribute.Int("modules.success", successCount),
		attribute.Int("modules.failed", failCount),
	)

	slog.InfoContext(ctx, "Module sync completed",
		"total", len(modules),
		"success", successCount,
		"failed", failCount)

	return nil
}
