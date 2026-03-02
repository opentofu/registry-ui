// Package syncproviders manages the sync-providers CLI command
package syncproviders

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
	"github.com/opentofu/registry-ui/pkg/provider"
	"github.com/opentofu/registry-ui/pkg/registry"
	"github.com/opentofu/registry-ui/pkg/telemetry"
)

func NewCommand() *cli.Command {
	return &cli.Command{
		Name:  "sync-providers",
		Usage: "Sync multiple providers matching a filter pattern from the registry",
		Flags: []cli.Flag{
			&cli.StringFlag{
				Name:     "filter",
				Aliases:  []string{"f"},
				Usage:    "Provider filter pattern (e.g., 'hashicorp/*', '*/aws', 'hashicorp/aws'). Supports wildcards.",
				Required: false,
			},
			&cli.StringFlag{
				Name:     "version",
				Aliases:  []string{"v"},
				Usage:    "Specific version to sync across all matching providers (e.g., 6.7.0). If not provided, syncs all missing versions",
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
	ctx, span := telemetry.Tracer().Start(ctx, "cmd.sync_providers")
	defer span.End()

	// Generate batch job ID for correlating all provider traces
	batchJobID := uuid.New().String()

	filter := cmd.String("filter")
	specificVersion := cmd.String("version")

	span.SetAttributes(
		attribute.String("filter", filter),
		attribute.String(telemetry.BatchJobIDKey, batchJobID),
		attribute.String(telemetry.BatchJobNameKey, "sync-providers"),
	)

	if specificVersion != "" {
		span.SetAttributes(attribute.String("version", specificVersion))
		slog.InfoContext(ctx, "Syncing providers with specific version from registry",
			"filter", filter, "version", specificVersion)
	} else {
		slog.InfoContext(ctx, "Syncing providers from registry",
			"filter", filter)
	}

	// Create provider reader
	providerReader, err := provider.NewProviderReader(cfg)
	if err != nil {
		span.RecordError(err)
		span.SetStatus(codes.Error, err.Error())
		slog.ErrorContext(ctx, "Failed to create provider reader", "error", err)
		return fmt.Errorf("failed to create provider reader: %w", err)
	}

	// Create registry client to list providers
	registryClient, err := registry.New(cfg.RegistryPath)
	if err != nil {
		span.RecordError(err)
		span.SetStatus(codes.Error, err.Error())
		return fmt.Errorf("failed to create registry client: %w", err)
	}

	// Update registry to ensure it's cloned and up-to-date
	err = registryClient.Update(ctx)
	if err != nil {
		span.RecordError(err)
		span.SetStatus(codes.Error, err.Error())
		return fmt.Errorf("failed to update registry: %w", err)
	}

	// Get list of providers matching filter
	providers, err := registryClient.ListProviders(ctx, filter)
	if err != nil {
		span.RecordError(err)
		span.SetStatus(codes.Error, err.Error())
		return fmt.Errorf("failed to list providers with filter %s: %w", filter, err)
	}

	if len(providers) == 0 {
		slog.InfoContext(ctx, "No providers found matching filter", "filter", filter)
		return nil
	}

	slog.InfoContext(ctx, "Found providers to sync", "count", len(providers), "filter", filter)

	// Sync each provider in parallel using errgroup
	g, gctx := errgroup.WithContext(ctx)
	g.SetLimit(cfg.Concurrency.Provider)

	slog.DebugContext(ctx, "Starting parallel provider sync",
		"providers", len(providers),
		"concurrency", cfg.Concurrency.Provider)

	var (
		successCount int
		failCount    int
		mu           sync.Mutex
	)

	for _, prov := range providers {
		g.Go(func() error {
			// Start a new trace for each provider, linked back to the batch job span
			provCtx, provSpan := telemetry.LinkedSpanStart(gctx, fmt.Sprintf("provider.sync.%s/%s", prov.Namespace, prov.Name),
				trace.WithAttributes(
					attribute.String(telemetry.BatchJobIDKey, batchJobID),
					attribute.String(telemetry.BatchJobNameKey, "sync-providers"),
					attribute.String("provider.namespace", prov.Namespace),
					attribute.String("provider.name", prov.Name),
				),
			)
			defer provSpan.End()

			logger := slog.With("provider.namespace", prov.Namespace, "provider.name", prov.Name)

			logger.InfoContext(provCtx, "Syncing provider")
			// Sync the provider version(s)
			var syncErr error
			if specificVersion != "" {
				syncErr = providerReader.EnsureParentRecords(provCtx, &prov)
				if syncErr == nil {
					_, syncErr = providerReader.IndexVersion(provCtx, &prov, specificVersion)
				}
				if syncErr == nil {
					// Regenerate per-provider version index (IndexAllVersions does this
					// internally, but IndexVersion does not)
					providerReader.RegenerateProviderVersionIndex(provCtx, prov.Namespace, prov.Name)
				}
			} else {
				syncErr = providerReader.ScrapeAllVersions(provCtx, &prov)
			}

			mu.Lock()
			defer mu.Unlock()

			if syncErr != nil {
				logger.ErrorContext(provCtx, "Failed to sync provider", "error", syncErr)
				provSpan.RecordError(syncErr)
				failCount++
			} else {
				successCount++
			}

			return nil // Don't fail the entire operation if one provider fails
		})
	}

	// Wait for all providers to complete
	if err := g.Wait(); err != nil {
		span.RecordError(err)
		span.SetStatus(codes.Error, err.Error())
		return fmt.Errorf("provider sync failed: %w", err)
	}

	span.SetAttributes(
		attribute.Int("providers.total", len(providers)),
		attribute.Int("providers.success", successCount),
		attribute.Int("providers.failed", failCount),
	)

	slog.InfoContext(ctx, "Provider sync completed",
		"total", len(providers),
		"success", successCount,
		"failed", failCount)

	return nil
}
