package index

import (
	"context"
	"fmt"
	"log/slog"
	"path/filepath"
	"strings"
	"sync"
	"time"

	"github.com/jackc/pgx/v5"
	"go.opentelemetry.io/otel/attribute"
	"golang.org/x/sync/errgroup"

	"github.com/opentofu/registry-ui/pkg/git"
	"github.com/opentofu/registry-ui/pkg/license"
	"github.com/opentofu/registry-ui/pkg/providers"
	"github.com/opentofu/registry-ui/pkg/registry"
)

// getExistingVersions queries the database for versions that already exist for a provider
func (s *IndexService) getExistingVersions(ctx context.Context, namespace, name string) ([]string, error) {
	ctx, span := s.tracer.Start(ctx, "getExistingVersions")
	defer span.End()

	span.SetAttributes(
		attribute.String("provider.namespace", namespace),
		attribute.String("provider.name", name),
	)

	query := `SELECT version FROM provider_versions WHERE provider_namespace = $1 AND provider_name = $2 ORDER BY version`
	rows, err := s.pool.Query(ctx, query, namespace, name)
	if err != nil {
		span.RecordError(err)
		return nil, fmt.Errorf("failed to query existing versions: %w", err)
	}
	defer rows.Close()

	existingVersions, err := pgx.CollectRows(rows, pgx.RowTo[string])
	if err != nil {
		span.RecordError(err)
		return nil, fmt.Errorf("failed to collect versions: %w", err)
	}

	span.SetAttributes(attribute.Int("existing_versions.count", len(existingVersions)))
	slog.InfoContext(ctx, "Found existing versions in database",
		"provider", fmt.Sprintf("%s/%s", namespace, name),
		"count", len(existingVersions))

	return existingVersions, nil
}

// calculateMissingVersions returns versions that exist in registry but not in database
func calculateMissingVersions(registryVersions, existingVersions []string) []string {
	// Create a set of existing versions for fast lookup, normalizing format too to remove v
	existingSet := make(map[string]bool, len(existingVersions))
	for _, version := range existingVersions {
		normalizedVersion := version
		if strings.HasPrefix(version, "v") {
			normalizedVersion = strings.TrimPrefix(version, "v")
		}
		existingSet[normalizedVersion] = true
	}

	// Find versions in registry that are not in db
	var missingVersions []string
	for _, version := range registryVersions {
		if !existingSet[version] {
			missingVersions = append(missingVersions, version)
		}
	}

	return missingVersions
}

// prepareProviderRepo clones and updates a git repository for provider processing
func (s *IndexService) prepareProviderRepo(ctx context.Context, namespace, name string) (*git.Repo, error) {
	ctx, span := s.tracer.Start(ctx, "prepareProviderRepo")
	defer span.End()

	repoURL := fmt.Sprintf("https://github.com/%s/terraform-provider-%s", namespace, name)
	localPath := filepath.Join(s.config.WorkDir, "providers", namespace, name)

	span.SetAttributes(
		attribute.String("git.repo_url", repoURL),
		attribute.String("git.local_path", localPath),
		attribute.String("provider.namespace", namespace),
		attribute.String("provider.name", name),
	)

	slog.InfoContext(ctx, "Preparing repository for provider sync",
		"provider", fmt.Sprintf("%s/%s", namespace, name),
		"repo_url", repoURL,
		"local_path", localPath)

	repo, err := git.GetRepo(repoURL, localPath)
	if err != nil {
		span.RecordError(err)
		return nil, fmt.Errorf("failed to get repository: %w", err)
	}

	if err := repo.Clone(ctx); err != nil {
		span.RecordError(err)
		return nil, fmt.Errorf("failed to clone repository: %w", err)
	}

	// Update repository to get all latest tags
	if err := repo.Update(ctx); err != nil {
		span.RecordError(err)
		return nil, fmt.Errorf("failed to update repository: %w", err)
	}

	slog.InfoContext(ctx, "Successfully prepared repository",
		"provider", fmt.Sprintf("%s/%s", namespace, name))

	return repo, nil
}

// checkProviderExists verifies that a provider exists in the registry and returns its version data
func (s *IndexService) checkProviderExists(ctx context.Context, reg *registry.Registry, namespace, name string) ([]string, error) {
	ctx, span := s.tracer.Start(ctx, "checkProviderExists")
	defer span.End()

	span.SetAttributes(
		attribute.String("provider.namespace", namespace),
		attribute.String("provider.name", name),
	)

	slog.InfoContext(ctx, "Checking if provider exists in registry",
		"namespace", namespace, "name", name)

	// Get provider data from registry
	providerData, err := reg.GetProvider(namespace, name)
	if err != nil {
		span.RecordError(err)
		return nil, fmt.Errorf("failed to get provider from registry: %w", err)
	}

	registryVersions := providerData.Versions
	slog.InfoContext(ctx, "Found provider in registry", "namespace", namespace, "name", name, "versions_count", len(registryVersions))

	span.SetAttributes(attribute.Int("versions.registry_count", len(registryVersions)))
	return registryVersions, nil
}

// analyzeVersions compares registry versions against database versions
func (s *IndexService) analyzeVersions(ctx context.Context, namespace, name string, registryVersions []string) ([]string, []string, error) {
	ctx, span := s.tracer.Start(ctx, "analyzeVersions")
	defer span.End()

	span.SetAttributes(
		attribute.String("provider.namespace", namespace),
		attribute.String("provider.name", name),
		attribute.Int("versions.registry_total", len(registryVersions)),
	)

	slog.InfoContext(ctx, "Analyzing versions against database",
		"namespace", namespace, "name", name, "registry_versions", len(registryVersions))

	// Get existing versions from database
	existingVersions, err := s.getExistingVersions(ctx, namespace, name)
	if err != nil {
		// It's OK if the provider doesn't exist in DB yet, we'll create it later if needed
		slog.DebugContext(ctx, "Provider not found in database (this is OK for new providers)",
			"provider", fmt.Sprintf("%s/%s", namespace, name), "error", err)
		existingVersions = []string{} // Empty list, all registry versions will be considered missing
	}

	// Calculate missing versions
	missingVersions := calculateMissingVersions(registryVersions, existingVersions)

	span.SetAttributes(
		attribute.Int("versions.existing", len(existingVersions)),
		attribute.Int("versions.missing", len(missingVersions)),
	)

	slog.InfoContext(ctx, "Version analysis complete",
		"registry_total", len(registryVersions),
		"existing", len(existingVersions),
		"missing", len(missingVersions))

	return existingVersions, missingVersions, nil
}

// storeVersionInDatabase handles all database operations for a successfully processed version
func (s *IndexService) storeVersionInDatabase(ctx context.Context, tx pgx.Tx, namespace, name, version string, licenses license.List, tagDate time.Time) error {
	ctx, span := s.tracer.Start(ctx, "storeVersionInDatabase")
	defer span.End()

	span.SetAttributes(
		attribute.String("provider.namespace", namespace),
		attribute.String("provider.name", name),
		attribute.String("provider.version", version),
		attribute.Int("licenses.count", len(licenses)),
	)

	licenseAccepted := true

	// Store provider version
	versionQuery := `
		INSERT INTO provider_versions (provider_namespace, provider_name, version, license_accepted, tag_created_at)
		VALUES ($1, $2, $3, $4, $5)
		ON CONFLICT (provider_namespace, provider_name, version)
		DO UPDATE SET 
			license_accepted = EXCLUDED.license_accepted,
			tag_created_at = EXCLUDED.tag_created_at,
			updated_at = NOW()
	`
	if _, err := tx.Exec(ctx, versionQuery, namespace, name, version, licenseAccepted, tagDate); err != nil {
		span.RecordError(err)
		slog.ErrorContext(ctx, "Failed to store provider version in database",
			"version", version, "error", err)
		return fmt.Errorf("failed to store provider version %s in database: %w", version, err)
	}

	// Store detailed license information
	if len(licenses) > 0 {
		for _, lic := range licenses {
			// Insert license if it doesn't exist
			licenseQuery := `
				INSERT INTO licenses (spdx_id, name, redistributable)
				VALUES ($1, $2, $3)
				ON CONFLICT (spdx_id) DO UPDATE SET
					name = EXCLUDED.name,
					redistributable = EXCLUDED.redistributable,
					updated_at = NOW()
			`
			if _, err := tx.Exec(ctx, licenseQuery, lic.SPDX, lic.SPDX, lic.IsCompatible); err != nil {
				span.RecordError(err)
				slog.ErrorContext(ctx, "Failed to store license in database",
					"license", lic.SPDX, "error", err)
				return fmt.Errorf("failed to store license %s in database: %w", lic.SPDX, err)
			}

			// Insert provider version license relationship
			pvlQuery := `
				INSERT INTO provider_version_licenses (
					provider_namespace, provider_name, version, license_spdx_id, 
					confidence_score, file_path, match_type
				)
				VALUES ($1, $2, $3, $4, $5, $6, $7)
				ON CONFLICT (provider_namespace, provider_name, version, license_spdx_id, file_path) 
				DO UPDATE SET
					confidence_score = EXCLUDED.confidence_score,
					match_type = EXCLUDED.match_type
			`
			if _, err := tx.Exec(ctx, pvlQuery,
				namespace, name, version, lic.SPDX,
				lic.Confidence, lic.File, "detector"); err != nil {
				slog.WarnContext(ctx, "Failed to store provider version license relationship",
					"license", lic.SPDX, "error", err)
				continue
			}
		}
	}

	slog.DebugContext(ctx, "Stored provider version and licenses in database",
		"version", version, "license_accepted", licenseAccepted, "licenses_count", len(licenses))

	return nil
}

// processVersionSync processes a single version (checkout, license detection, doc scraping, database storage)
func (s *IndexService) processVersionSync(ctx context.Context, provider *providers.Provider, namespace, name, version string) VersionResult {
	start := time.Now()
	ctx, span := s.tracer.Start(ctx, "processVersionSync")
	defer span.End()

	span.SetAttributes(
		attribute.String("provider.namespace", namespace),
		attribute.String("provider.name", name),
		attribute.String("provider.version", version),
	)

	slog.DebugContext(ctx, "Starting processing for version", "version", version)
	result := VersionResult{Version: version}

	span.SetAttributes(attribute.Bool("version.already_exists", false))
	directory, cleanup, err := provider.CheckoutVersionForScraping(ctx, version)
	if err != nil {
		span.RecordError(err)
		slog.ErrorContext(ctx, "Failed to checkout version",
			"version", version, "error", err)
		result.Error = fmt.Errorf("failed to checkout version %s: %w", version, err)
		result.Duration = time.Since(start)
		return result
	}
	defer cleanup()

	licenses, err := provider.DetectLicensesInDirectory(ctx, directory)
	if err != nil {
		span.RecordError(err)
		slog.WarnContext(ctx, "Failed to detect licenses for version",
			"version", version, "error", err)
		result.Error = err
		result.LicensesStr = "N/A"
		result.Explanation = "Failed to detect licenses: " + err.Error()
		result.Duration = time.Since(start)
		return result
	}

	result.Licenses = licenses
	result.Redistributable = licenses.IsRedistributable()

	span.SetAttributes(
		attribute.Bool("license.redistributable", result.Redistributable),
		attribute.Int("license.count", len(licenses)),
	)

	result.LicensesStr = licenses.String()
	if result.LicensesStr == "" {
		result.LicensesStr = "None detected"
	}

	result.Explanation = licenses.Explain()
	slog.DebugContext(ctx, "Completed license detection for version",
		"version", version, "licenses", result.LicensesStr, "redistributable", result.Redistributable)

	// Get tag creation date for this version
	tagDate, err := provider.GetTagCreationDate(ctx, version)
	if err != nil {
		slog.ErrorContext(ctx, "Failed to get tag date for version", "version", version, "error", err)
		span.RecordError(err)
		result.Error = err
		result.Duration = time.Since(start)
		return result
	}

	slog.DebugContext(ctx, "Retrieved tag date for version", "version", version, "tag_date", tagDate)

	if result.Redistributable && result.Error == nil {
		// Begin transaction for this version's database operations
		tx, err := s.pool.Begin(ctx)
		if err != nil {
			span.RecordError(err)
			result.Error = fmt.Errorf("failed to begin transaction for version %s: %w", version, err)
			result.Duration = time.Since(start)
			return result
		}
		defer tx.Rollback(ctx) // Safe to call on committed transactions

		// Store provider version in database (required for foreign key)
		if err := s.storeVersionInDatabase(ctx, tx, namespace, name, version, result.Licenses, *tagDate); err != nil {
			span.RecordError(err)
			result.Error = err
			result.Duration = time.Since(start)
			return result
		}

		// Scrape and store documentation
		span.SetAttributes(attribute.Bool("docs.scraping", true))
		if err := s.scraper.ScrapeAndStore(ctx, namespace, name, version, directory, result.Licenses, tx); err != nil {
			span.RecordError(err)
			slog.ErrorContext(ctx, "Failed to scrape documentation", "version", version, "error", err)
			result.Error = fmt.Errorf("failed to scrape documentation for version %s: %w", version, err)
			result.Duration = time.Since(start)
			return result
		}

		// Commit transaction only if everything succeeded
		if err := tx.Commit(ctx); err != nil {
			span.RecordError(err)
			result.Error = fmt.Errorf("failed to commit transaction for version %s: %w", version, err)
			result.Duration = time.Since(start)
			return result
		}

		span.SetAttributes(attribute.Bool("docs.scraped", true))
		slog.DebugContext(ctx, "Successfully scraped documentation", "version", version)
	} else {
		span.SetAttributes(attribute.Bool("docs.scraping", false))
	}

	result.Duration = time.Since(start)
	return result
}

// processVersionsInParallel handles the parallel coordination and results collection
func (s *IndexService) processVersionsInParallel(ctx context.Context, provider *providers.Provider, namespace, name string, missingVersions []string) ([]VersionResult, error) {
	ctx, span := s.tracer.Start(ctx, "processVersionsInParallel")
	defer span.End()

	span.SetAttributes(
		attribute.String("provider.namespace", namespace),
		attribute.String("provider.name", name),
		attribute.Int("versions.missing", len(missingVersions)),
		attribute.Int("concurrency.limit", s.config.Concurrency.Version),
	)

	slog.InfoContext(ctx, "Processing missing versions in parallel",
		"count", len(missingVersions), "concurrency_limit", s.config.Concurrency.Version)

	// Process versions in parallel using errgroup
	results := make([]VersionResult, len(missingVersions))
	var mu sync.Mutex

	g, gctx := errgroup.WithContext(ctx)
	g.SetLimit(s.config.Concurrency.Version)

	processedCount := 0
	totalCount := len(missingVersions)
	docsScrapedCount := 0

	for i, version := range missingVersions {
		i, version := i, version // capture loop variables
		g.Go(func() error {
			// Process this version
			result := s.processVersionSync(gctx, provider, namespace, name, version)

			// Thread-safe write to results slice and progress tracking
			mu.Lock()
			results[i] = result
			processedCount++
			currentProgress := processedCount
			if result.Redistributable && result.Error == nil {
				docsScrapedCount++
			}
			currentDocsScraped := docsScrapedCount
			mu.Unlock()

			// Log progress every 5 completed versions or on completion
			if currentProgress%5 == 0 || currentProgress == totalCount {
				slog.InfoContext(gctx, "Processing progress",
					"completed", currentProgress, "total", totalCount,
					"docs_scraped", currentDocsScraped,
					"percentage", fmt.Sprintf("%.1f%%", float64(currentProgress)/float64(totalCount)*100))
			}

			// Don't fail entire batch on single version error
			return nil
		})
	}

	// Wait for all goroutines to complete
	if err := g.Wait(); err != nil {
		span.RecordError(err)
		slog.ErrorContext(ctx, "Error during parallel processing", "error", err)
		return nil, fmt.Errorf("error during parallel processing: %w", err)
	}

	span.SetAttributes(
		attribute.Int("versions.processed", len(missingVersions)),
		attribute.Int("docs.scraped", docsScrapedCount),
	)

	slog.InfoContext(ctx, "Completed parallel processing for missing versions",
		"total_processed", len(missingVersions), "docs_scraped", docsScrapedCount)

	return results, nil
}

// ensureProviderWithAliasInDB ensures that the provider and its repository exist in the database with alias information
func (s *IndexService) ensureProviderWithAliasInDB(ctx context.Context, originalNamespace, originalName, canonicalNamespace, canonicalName string, isAlias bool) error {
	ctx, span := s.tracer.Start(ctx, "ensureProviderWithAliasInDB")
	defer span.End()

	span.SetAttributes(
		attribute.String("provider.original_namespace", originalNamespace),
		attribute.String("provider.original_name", originalName),
		attribute.String("provider.canonical_namespace", canonicalNamespace),
		attribute.String("provider.canonical_name", canonicalName),
		attribute.Bool("provider.is_alias", isAlias),
	)

	// Repository info - use canonical address for repository reference since that's where the code lives
	repoOrg := canonicalNamespace
	repoName := fmt.Sprintf("terraform-provider-%s", canonicalName)

	// First, ensure the repository exists (using canonical address)
	repoQuery := `
		INSERT INTO repositories (organisation, name) 
		VALUES ($1, $2) 
		ON CONFLICT (organisation, name) DO NOTHING`

	_, err := s.pool.Exec(ctx, repoQuery, repoOrg, repoName)
	if err != nil {
		span.RecordError(err)
		return fmt.Errorf("failed to ensure repository %s/%s exists in database: %w", repoOrg, repoName, err)
	}

	// Store alias information in the alias field if this is an alias
	var aliasValue *string
	if isAlias {
		aliasAddr := fmt.Sprintf("%s/%s", canonicalNamespace, canonicalName)
		aliasValue = &aliasAddr
	}

	// Then, ensure the provider exists under the original (requested) address
	providerQuery := `
		INSERT INTO providers (namespace, name, alias, repo_organisation, repo_name) 
		VALUES ($1, $2, $3, $4, $5) 
		ON CONFLICT (namespace, name) DO UPDATE SET 
			alias = EXCLUDED.alias,
			repo_organisation = EXCLUDED.repo_organisation,
			repo_name = EXCLUDED.repo_name`

	_, err = s.pool.Exec(ctx, providerQuery, originalNamespace, originalName, aliasValue, repoOrg, repoName)
	if err != nil {
		span.RecordError(err)
		return fmt.Errorf("failed to ensure provider %s/%s exists in database: %w", originalNamespace, originalName, err)
	}

	if isAlias {
		slog.InfoContext(ctx, "Ensured provider with alias exists in database",
			"requested", fmt.Sprintf("%s/%s", originalNamespace, originalName),
			"canonical", fmt.Sprintf("%s/%s", canonicalNamespace, canonicalName),
			"repository", fmt.Sprintf("%s/%s", repoOrg, repoName))
	} else {
		slog.InfoContext(ctx, "Ensured provider exists in database",
			"provider", fmt.Sprintf("%s/%s", originalNamespace, originalName),
			"repository", fmt.Sprintf("%s/%s", repoOrg, repoName))
	}

	return nil
}

// processProvidersInParallel processes multiple providers in parallel with configured concurrency
func (s *IndexService) processProvidersInParallel(ctx context.Context, reg *registry.Registry, providers []registry.Provider, version string) (*MultiProviderIndexResponse, error) {
	ctx, span := s.tracer.Start(ctx, "processProvidersInParallel")
	defer span.End()

	span.SetAttributes(
		attribute.Int("providers.total", len(providers)),
		attribute.Int("concurrency.limit", s.config.Concurrency.Provider),
	)

	if version != "" {
		span.SetAttributes(attribute.String("version", version))
	}

	slog.InfoContext(ctx, "Processing multiple providers in parallel",
		"count", len(providers), "concurrency_limit", s.config.Concurrency.Provider, "version", version)

	// Process providers in parallel using errgroup
	results := make([]ProviderIndexResult, len(providers))
	var mu sync.Mutex

	g, gctx := errgroup.WithContext(ctx)
	g.SetLimit(s.config.Concurrency.Provider)

	for i, provider := range providers {
		i, provider := i, provider // capture loop variables
		g.Go(func() error {
			// Process this provider with pre-prepared registry
			result := s.processProviderSyncWithRegistry(gctx, reg, provider.Namespace, provider.Name, version)

			// Thread-safe write to results slice and progress tracking
			mu.Lock()
			results[i] = result

			mu.Unlock()

			// Don't fail entire batch on single provider error
			return nil
		})
	}

	// Wait for all goroutines to complete
	if err := g.Wait(); err != nil {
		span.RecordError(err)
		slog.ErrorContext(ctx, "Error during parallel provider processing", "error", err)
		return nil, fmt.Errorf("error during parallel provider processing: %w", err)
	}

	// Build response
	response := &MultiProviderIndexResponse{
		TotalProviders:  len(providers),
		ProviderResults: results,
	}

	// Count successes and failures
	processedProviders := 0
	failedProviders := 0
	for _, result := range results {
		if result.Error != nil {
			failedProviders++
		} else {
			processedProviders++
		}
	}

	response.ProcessedProviders = processedProviders
	response.FailedProviders = failedProviders

	span.SetAttributes(
		attribute.Int("providers.processed", processedProviders),
		attribute.Int("providers.failed", failedProviders),
	)

	slog.InfoContext(ctx, "Completed parallel processing for multiple providers",
		"total_providers", len(providers),
		"processed", processedProviders,
		"failed", failedProviders)

	return response, nil
}

// processProviderSyncWithRegistry processes a single provider using a pre-prepared registry
func (s *IndexService) processProviderSyncWithRegistry(ctx context.Context, reg *registry.Registry, namespace, name, version string) ProviderIndexResult {
	start := time.Now()
	ctx, span := s.tracer.Start(ctx, "processProviderSyncWithRegistry")
	defer span.End()

	span.SetAttributes(
		attribute.String("provider.namespace", namespace),
		attribute.String("provider.name", name),
	)

	if version != "" {
		span.SetAttributes(attribute.String("provider.version", version))
	}

	slog.DebugContext(ctx, "Starting processing for provider", "provider", fmt.Sprintf("%s/%s", namespace, name), "version", version)

	result := ProviderIndexResult{
		Namespace: namespace,
		Name:      name,
		Duration:  0,
	}

	// Call the internal method with pre-prepared registry
	response, err := s.indexProviderVersionWithRegistry(ctx, reg, namespace, name, version)
	if err != nil {
		span.RecordError(err)
		slog.ErrorContext(ctx, "Failed to index provider",
			"provider", fmt.Sprintf("%s/%s", namespace, name), "version", version, "error", err)
		result.Error = err
		result.Duration = time.Since(start)
		return result
	}

	result.Response = response
	result.Duration = time.Since(start)

	slog.DebugContext(ctx, "Successfully processed provider",
		"provider", fmt.Sprintf("%s/%s", namespace, name),
		"versions_processed", response.ProcessedVersions,
		"duration", result.Duration)

	return result
}
