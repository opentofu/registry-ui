package index

import (
	"context"
	"fmt"
	"log/slog"

	"go.opentelemetry.io/otel/attribute"
)

// IndexProviderVersion is the main entry point for indexing provider versions
// It processes a specific provider and optionally a specific version
func (s *IndexService) IndexProviderVersion(ctx context.Context, namespace, name, version string) (*IndexResponse, error) {
	ctx, span := s.tracer.Start(ctx, "index.IndexProviderVersion")
	defer span.End()

	span.SetAttributes(
		attribute.String("provider.namespace", namespace),
		attribute.String("provider.name", name),
	)

	// Resolve provider aliases to determine canonical repo for documentation scraping
	canonicalNamespace, canonicalName, isAlias, err := s.ResolveProviderAlias(ctx, namespace, name)
	if err != nil {
		span.RecordError(err)
		return nil, fmt.Errorf("failed to resolve provider alias: %w", err)
	}

	// Store original addresses for database operations
	originalNamespace, originalName := namespace, name
	
	if isAlias {
		span.SetAttributes(
			attribute.String("provider.canonical_namespace", canonicalNamespace),
			attribute.String("provider.canonical_name", canonicalName),
			attribute.Bool("provider.is_alias", true),
		)
		slog.InfoContext(ctx, "Resolved provider alias",
			"requested", fmt.Sprintf("%s/%s", originalNamespace, originalName),
			"canonical", fmt.Sprintf("%s/%s", canonicalNamespace, canonicalName),
			"note", "Will store under requested address but scrape from canonical repo")
	}

	if version != "" {
		span.SetAttributes(attribute.String("provider.version", version))
		slog.InfoContext(ctx, "Indexing specific provider version",
			"requested", fmt.Sprintf("%s/%s", originalNamespace, originalName), "version", version)
	} else {
		slog.InfoContext(ctx, "Indexing all provider versions",
			"requested", fmt.Sprintf("%s/%s", originalNamespace, originalName))
	}

	// Prepare registry
	reg, err := s.prepareRegistry(ctx)
	if err != nil {
		span.RecordError(err)
		return nil, err
	}

	// Check provider exists in registry and get versions (use canonical address for registry lookup)
	registryVersions, err := s.checkProviderExists(ctx, reg, canonicalNamespace, canonicalName)
	if err != nil {
		span.RecordError(err)
		return nil, err
	}

	// Filter to specific version if provided
	if version != "" {
		found := false
		for _, v := range registryVersions {
			if v == version {
				found = true
				break
			}
		}
		if !found {
			return nil, fmt.Errorf("version %s not found in registry for provider %s/%s", version, canonicalNamespace, canonicalName)
		}
		registryVersions = []string{version}
	}

	// Analyze versions against database (use original address for database lookup)
	existingVersions, missingVersions, err := s.analyzeVersions(ctx, originalNamespace, originalName, registryVersions)
	if err != nil {
		span.RecordError(err)
		return nil, err
	}

	span.SetAttributes(
		attribute.Int("versions.registry_total", len(registryVersions)),
		attribute.Int("versions.existing", len(existingVersions)),
		attribute.Int("versions.missing", len(missingVersions)),
	)

	response := &IndexResponse{
		TotalVersions:     len(registryVersions),
		ProcessedVersions: 0,
		SkippedVersions:   len(existingVersions),
		FailedVersions:    0,
		Results:           []VersionResult{},
	}

	if len(missingVersions) == 0 {
		return response, nil
	}

	// Prepare provider using canonical address for repository operations but pass both addresses
	provider, err := s.prepareProviderWithAlias(ctx, originalNamespace, originalName, canonicalNamespace, canonicalName, isAlias)
	if err != nil {
		span.RecordError(err)
		return nil, err
	}

	results, err := s.processVersionsInParallel(ctx, provider, originalNamespace, originalName, missingVersions)
	if err != nil {
		span.RecordError(err)
		return nil, err
	}

	response.Results = results
	response.ProcessedVersions = len(results)

	// Count failures
	failureCount := 0
	for _, result := range results {
		if result.Error != nil {
			failureCount++
		}
	}
	response.FailedVersions = failureCount

	slog.InfoContext(ctx, "Successfully completed provider indexing",
		"requested", fmt.Sprintf("%s/%s", originalNamespace, originalName),
		"canonical", fmt.Sprintf("%s/%s", canonicalNamespace, canonicalName),
		"total_versions", response.TotalVersions,
		"processed", response.ProcessedVersions,
		"skipped", response.SkippedVersions,
		"failed", response.FailedVersions)

	return response, nil
}

// IndexMultipleProviders indexes multiple providers matching the given filter pattern
func (s *IndexService) IndexMultipleProviders(ctx context.Context, filter, version string) (*MultiProviderIndexResponse, error) {
	ctx, span := s.tracer.Start(ctx, "index.IndexMultipleProviders")
	defer span.End()

	span.SetAttributes(attribute.String("filter", filter))
	if version != "" {
		span.SetAttributes(attribute.String("version", version))
	}

	slog.InfoContext(ctx, "Starting multiple provider indexing", "filter", filter, "version", version)

	// Prepare registry to get list of providers
	reg, err := s.prepareRegistry(ctx)
	if err != nil {
		span.RecordError(err)
		return nil, err
	}

	// Get list of providers matching filter
	providers, err := reg.ListProviders(filter)
	if err != nil {
		span.RecordError(err)
		return nil, fmt.Errorf("failed to list providers with filter %s: %w", filter, err)
	}

	if len(providers) == 0 {
		return &MultiProviderIndexResponse{
			TotalProviders:     0,
			ProcessedProviders: 0,
			SkippedProviders:   0,
			FailedProviders:    0,
			ProviderResults:    []ProviderIndexResult{},
		}, nil
	}

	span.SetAttributes(attribute.Int("providers.total", len(providers)))
	slog.InfoContext(ctx, "Found providers to index", "count", len(providers), "filter", filter)

	// Process providers in parallel with configured concurrency
	return s.processProvidersInParallel(ctx, providers, version)
}
