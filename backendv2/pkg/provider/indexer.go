package provider

import (
	"context"
	"fmt"
	"log/slog"
	"path/filepath"
	"slices"
	"sync"
	"time"

	"github.com/jackc/pgx/v5"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/codes"
	"golang.org/x/sync/errgroup"

	"github.com/opentofu/registry-ui/pkg/git"
	"github.com/opentofu/registry-ui/pkg/index"
	"github.com/opentofu/registry-ui/pkg/license"
	"github.com/opentofu/registry-ui/pkg/provider/scraper"
	"github.com/opentofu/registry-ui/pkg/provider/storage"
	"github.com/opentofu/registry-ui/pkg/registry"
	"github.com/opentofu/registry-ui/pkg/repository"
	"github.com/opentofu/registry-ui/pkg/telemetry"
)

// IndexVersion indexes a specific version of a provider
// provider parameter must be provided by the caller to avoid redundant file reads
func (p *ProviderReader) IndexVersion(ctx context.Context, provider *registry.Provider, version string) (response *IndexResponse, err error) {
	ctx, span := telemetry.Tracer().Start(ctx, "provider.IndexVersion")
	defer func() {
		if err != nil {
			span.SetStatus(codes.Error, err.Error())
		}
		span.End()
	}()

	namespace := provider.Namespace
	name := provider.Name

	span.SetAttributes(
		attribute.String("provider.namespace", namespace),
		attribute.String("provider.name", name),
		attribute.String("provider.version", version),
	)

	slog.DebugContext(ctx, "Starting provider version indexing",
		"provider", fmt.Sprintf("%s/%s", namespace, name), "version", version)

	// Get the versions from the provider
	versions := provider.Versions

	versionExists := slices.Contains(versions, version)

	if !versionExists {
		return nil, fmt.Errorf("version %s not found for provider %s/%s", version, namespace, name)
	}

	// Checkout the version for scraping
	var workDir string
	var cleanup func()
	workDir, cleanup, err = p.CheckoutVersionForScraping(ctx, namespace, name, version)
	if err != nil {
		span.RecordError(err)
		span.SetStatus(codes.Error, err.Error())
		return nil, fmt.Errorf("failed to checkout version: %w", err)
	}
	defer cleanup()

	// Detect licenses first
	var licenses license.List
	licenses, err = p.DetectLicensesInDirectory(ctx, namespace, name, workDir)
	if err != nil {
		span.RecordError(err)
		span.SetStatus(codes.Error, err.Error())
		return nil, fmt.Errorf("failed to detect licenses: %w", err)
	}

	// Determine if licenses are acceptable for documentation scraping
	// No license (nil/empty) or incompatible licenses = skip documentation but store version
	var licenseAccepted bool

	if len(licenses) == 0 {
		licenseAccepted = false
		slog.WarnContext(ctx, "Provider has no license, will store version but skip documentation",
			"provider", fmt.Sprintf("%s/%s", namespace, name),
			"version", version)
		span.SetAttributes(
			attribute.String("provider.skip_reason", "no_license"),
			attribute.Bool("provider.docs_skipped", true),
		)
	} else if licenses.HasIncompatible() {
		licenseAccepted = false
		incompatibleList := licenses.String()
		slog.WarnContext(ctx, "Provider has incompatible license(s), will store version but skip documentation",
			"provider", fmt.Sprintf("%s/%s", namespace, name),
			"version", version,
			"licenses", incompatibleList)
		span.SetAttributes(
			attribute.String("provider.skip_reason", "incompatible_license"),
			attribute.Bool("provider.docs_skipped", true),
			attribute.String("provider.incompatible_licenses", incompatibleList),
		)
	} else {
		licenseAccepted = true
	}

	// Create documentation scraper
	docScraper := scraper.New(p.config, p.s3Client, p.db)

	// Initialize doc count and docs
	var docCount int
	var docs map[string]*scraper.DocItem

	// Only scrape documentation if license is acceptable
	if licenseAccepted {
		// Get documentation to count it
		docs, err = docScraper.ScrapeDocumentation(ctx, namespace, name, version, workDir)
		if err != nil {
			span.RecordError(err)
			span.SetStatus(codes.Error, err.Error())
			return nil, fmt.Errorf("failed to scrape documentation: %w", err)
		}
		docCount = len(docs)
	} else {
		docCount = 0
		slog.InfoContext(ctx, "Skipping documentation scraping due to incompatible license",
			"provider", fmt.Sprintf("%s/%s", namespace, name), "version", version)
	}

	// Start a database transaction for atomic operations
	// Note: Repository and provider records are stored BEFORE parallel processing
	// in IndexAllVersions to avoid row lock contention between concurrent versions
	var tx pgx.Tx
	tx, err = p.db.Begin(ctx)
	if err != nil {
		span.RecordError(err)
		span.SetStatus(codes.Error, err.Error())
		return nil, fmt.Errorf("failed to start transaction: %w", err)
	}
	defer tx.Rollback(ctx)

	// Get tag creation date
	var tagCreatedAt *time.Time
	tagCreatedAt, err = p.GetTagCreationDate(ctx, namespace, name, version)
	if err != nil {
		// Log warning but don't fail - tag date is not critical
		slog.WarnContext(ctx, "Failed to get tag creation date",
			"provider", fmt.Sprintf("%s/%s", namespace, name),
			"version", version, "error", err)
	}

	// Determine scrape status and skip reason
	var scrapeStatus, skipReason string
	if !licenseAccepted {
		scrapeStatus = "skipped"
		if len(licenses) == 0 {
			skipReason = "no_license"
		} else {
			skipReason = "incompatible_license"
		}
	} else {
		scrapeStatus = "completed"
		skipReason = ""
	}

	// Store provider version in database
	err = storage.StoreProviderVersion(ctx, tx, namespace, name, version, docCount, len(licenses), tagCreatedAt, licenseAccepted, scrapeStatus, skipReason, "")
	if err != nil {
		span.RecordError(err)
		span.SetStatus(codes.Error, err.Error())
		return nil, fmt.Errorf("failed to store provider version: %w", err)
	}

	// Store license information (always, for complete audit trail)
	if len(licenses) > 0 {
		err = storage.StoreProviderLicenses(ctx, tx, namespace, name, version, licenses)
		if err != nil {
			span.RecordError(err)
			span.SetStatus(codes.Error, err.Error())
			return nil, fmt.Errorf("failed to store provider licenses: %w", err)
		}
		slog.DebugContext(ctx, "Stored license information in database",
			"provider", fmt.Sprintf("%s/%s", namespace, name),
			"version", version,
			"license_count", len(licenses))
	}

	// Store documents and complete the scraping process only if license was accepted
	if licenseAccepted {
		err = docScraper.ScrapeAndStore(ctx, namespace, name, version, workDir, licenses, tx)
		if err != nil {
			span.RecordError(err)
			span.SetStatus(codes.Error, err.Error())
			return nil, fmt.Errorf("failed to store documentation: %w", err)
		}
	}

	err = tx.Commit(ctx)
	if err != nil {
		span.RecordError(err)
		span.SetStatus(codes.Error, err.Error())
		return nil, fmt.Errorf("failed to commit transaction: %w", err)
	}

	response = &IndexResponse{
		Namespace:      namespace,
		Name:           name,
		Version:        version,
		ProcessedAt:    time.Now().UTC(),
		DocumentsCount: docCount,
		LicensesCount:  len(licenses),
		Success:        true,
	}

	if licenseAccepted {
		slog.DebugContext(ctx, "Successfully indexed provider version with documentation",
			"provider", fmt.Sprintf("%s/%s", namespace, name),
			"version", version,
			"docs", docCount,
			"licenses", len(licenses))
	} else {
		slog.DebugContext(ctx, "Successfully indexed provider version (documentation skipped due to incompatible license)",
			"provider", fmt.Sprintf("%s/%s", namespace, name),
			"version", version,
			"docs", docCount,
			"licenses", len(licenses))
	}

	return response, nil
}

// IndexAllVersions indexes all versions of a provider
// registryProvider must be provided by the caller
func (p *ProviderReader) IndexAllVersions(ctx context.Context, namespace, name string, registryProvider *registry.Provider) ([]*IndexResponse, error) {
	if registryProvider == nil {
		panic("registryProvider cannot be nil")
	}

	ctx, span := telemetry.Tracer().Start(ctx, "provider.IndexAllVersions")
	defer span.End()

	span.SetAttributes(
		attribute.String("provider.namespace", namespace),
		attribute.String("provider.name", name),
	)

	provider := registryProvider

	// Get the versions from the provider
	allVersions := provider.Versions

	// Get existing versions from database to avoid redundant work
	existingVersions, err := storage.GetExistingProviderVersions(ctx, p.db, namespace, name)
	if err != nil {
		span.RecordError(err)
		span.SetStatus(codes.Error, err.Error())
		return nil, fmt.Errorf("failed to get existing versions: %w", err)
	}

	// Convert to map for fast lookup
	existingSet := make(map[string]bool)
	for _, version := range existingVersions {
		existingSet[version] = true
	}

	// Filter to only versions that need processing
	var versionsToProcess []string
	var skippedVersions []string
	for _, version := range allVersions {
		if existingSet[version] {
			skippedVersions = append(skippedVersions, version)
		} else {
			versionsToProcess = append(versionsToProcess, version)
		}
	}

	span.SetAttributes(
		attribute.Int("provider.versions_total", len(allVersions)),
		attribute.Int("provider.versions_existing", len(skippedVersions)),
		attribute.Int("provider.versions_to_process", len(versionsToProcess)),
	)

	slog.DebugContext(ctx, "Version processing analysis",
		"provider", fmt.Sprintf("%s/%s", namespace, name),
		"total_versions", len(allVersions),
		"existing_versions", len(skippedVersions),
		"to_process", len(versionsToProcess))

	var responses []*IndexResponse
	var responsesMu sync.Mutex
	var failedVersions []string
	var failedMu sync.Mutex

	// Add responses for skipped versions (already exist)
	for _, version := range skippedVersions {
		responses = append(responses, &IndexResponse{
			Namespace:      namespace,
			Name:           name,
			Version:        version,
			ProcessedAt:    time.Now().UTC(), // We don't query for exact time to keep it fast
			DocumentsCount: 0,                // We don't track counts
			LicensesCount:  0,                // We don't track counts
			Success:        true,
		})
	}

	// If no versions need processing, return early with skipped versions
	if len(versionsToProcess) == 0 {
		slog.DebugContext(ctx, "No versions to process, all versions already exist",
			"provider", fmt.Sprintf("%s/%s", namespace, name))
		return responses, nil
	}

	// Pre-clone and fetch tags before parallel processing to avoid race conditions
	repoURL := fmt.Sprintf("https://github.com/%s/terraform-provider-%s", namespace, name)
	localPath := filepath.Join(p.config.WorkDir, "providers", namespace, name)

	slog.DebugContext(ctx, "Preparing provider repository for parallel processing",
		"provider", fmt.Sprintf("%s/%s", namespace, name),
		"repo_url", repoURL,
		"local_path", localPath)

	repo, err := git.GetRepo(repoURL, localPath)
	if err != nil {
		span.RecordError(err)
		span.SetStatus(codes.Error, err.Error())
		return nil, fmt.Errorf("failed to get repository: %w", err)
	}

	// EnsureCloned the repository (if not already cloned)
	if err := repo.EnsureCloned(ctx); err != nil {
		span.RecordError(err)
		span.SetStatus(codes.Error, err.Error())
		return nil, fmt.Errorf("failed to clone repository: %w", err)
	}

	// Fetch all tags from remote
	if err := repo.FetchTags(ctx); err != nil {
		span.RecordError(err)
		span.SetStatus(codes.Error, err.Error())
		return nil, fmt.Errorf("failed to fetch tags: %w", err)
	}

	slog.InfoContext(ctx, "Successfully prepared repository for parallel processing",
		"provider", fmt.Sprintf("%s/%s", namespace, name))

	// Store repository and provider ONCE before parallel version processing
	// This eliminates row lock contention between concurrent versions
	repoOrg := namespace
	repoName := fmt.Sprintf("terraform-provider-%s", name)

	if err := storage.StoreRepository(ctx, p.db, repoOrg, repoName); err != nil {
		span.RecordError(err)
		span.SetStatus(codes.Error, err.Error())
		return nil, fmt.Errorf("failed to store repository: %w", err)
	}

	if err := storage.StoreProvider(ctx, p.db, namespace, name, repoOrg, repoName, registryProvider.Warnings); err != nil {
		span.RecordError(err)
		span.SetStatus(codes.Error, err.Error())
		return nil, fmt.Errorf("failed to store provider: %w", err)
	}

	slog.DebugContext(ctx, "Stored repository and provider metadata",
		"provider", fmt.Sprintf("%s/%s", namespace, name))

	// Process each version that needs processing in parallel using errgroup
	g, gctx := errgroup.WithContext(ctx)
	g.SetLimit(p.config.Concurrency.Version)

	slog.InfoContext(ctx, "Starting parallel version processing",
		"provider", fmt.Sprintf("%s/%s", namespace, name),
		"versions_to_process", len(versionsToProcess),
		"concurrency_limit", p.config.Concurrency.Version)

	for _, version := range versionsToProcess {
		g.Go(func() error {
			// Create a new span for this version processing
			versionCtx, versionSpan := telemetry.Tracer().Start(gctx, "provider.IndexAllVersions.version")
			defer versionSpan.End()

			versionSpan.SetAttributes(
				attribute.String("provider.namespace", namespace),
				attribute.String("provider.name", name),
				attribute.String("provider.version", version),
			)

			// Process the version with its own context and span
			// Pass the provider data to avoid redundant file reads
			response, err := p.IndexVersion(versionCtx, provider, version)
			if err != nil {
				versionSpan.RecordError(err)
				versionSpan.SetStatus(codes.Error, err.Error())
				slog.WarnContext(versionCtx, "Failed to index version, storing as failed to prevent retry",
					"provider", fmt.Sprintf("%s/%s", namespace, name),
					"version", version, "error", err)

				// Store failed version in database to prevent re-scraping
				// Pass the provider data to avoid redundant file reads
				storeErr := p.storeFailedVersion(versionCtx, namespace, name, version, err.Error(), provider)
				if storeErr != nil {
					slog.ErrorContext(versionCtx, "Failed to store failed version record",
						"provider", fmt.Sprintf("%s/%s", namespace, name),
						"version", version, "error", storeErr)
				}

				// Thread-safe append to failed versions
				failedMu.Lock()
				failedVersions = append(failedVersions, version)
				failedMu.Unlock()

				// Create error response
				response = &IndexResponse{
					Namespace:    namespace,
					Name:         name,
					Version:      version,
					ProcessedAt:  time.Now().UTC(),
					Success:      false,
					ErrorMessage: err.Error(),
				}
			}

			// Thread-safe append to responses
			responsesMu.Lock()
			responses = append(responses, response)
			responsesMu.Unlock()

			return nil // don't fail the entire group on one version failure
		})
	}

	// Wait for all versions to complete
	if err := g.Wait(); err != nil {
		span.RecordError(err)
		span.SetStatus(codes.Error, err.Error())
		return nil, fmt.Errorf("errgroup failed: %w", err)
	}

	if len(failedVersions) > 0 {
		slog.WarnContext(ctx, "Some versions failed to index",
			"provider", fmt.Sprintf("%s/%s", namespace, name),
			"failed_versions", failedVersions)
	}

	// Sync repository metadata from GitHub (stats, fork info, etc.)
	if p.githubClient != nil {
		repoOrg := namespace
		repoName := fmt.Sprintf("terraform-provider-%s", name)

		slog.DebugContext(ctx, "Syncing repository metadata from GitHub",
			"provider", fmt.Sprintf("%s/%s", namespace, name),
			"repository", fmt.Sprintf("%s/%s", repoOrg, repoName))

		err := repository.SyncRepositoryMetadata(ctx, p.db, p.githubClient, repoOrg, repoName)
		if err != nil {
			// Log error but don't fail the sync process
			slog.WarnContext(ctx, "Failed to sync repository metadata",
				"provider", fmt.Sprintf("%s/%s", namespace, name),
				"repository", fmt.Sprintf("%s/%s", repoOrg, repoName),
				"error", err)
		} else {
			slog.InfoContext(ctx, "Successfully synced repository metadata",
				"provider", fmt.Sprintf("%s/%s", namespace, name),
				"repository", fmt.Sprintf("%s/%s", repoOrg, repoName))
		}
	}

	// Generate and upload provider version index after processing versions
	slog.InfoContext(ctx, "Generating provider version index",
		"provider", fmt.Sprintf("%s/%s", namespace, name))

	providerIndex, err := index.GenerateProviderVersionIndex(ctx, p.db, namespace, name)
	if err != nil {
		slog.WarnContext(ctx, "Failed to generate provider version index",
			"provider", fmt.Sprintf("%s/%s", namespace, name),
			"error", err)
	} else {
		// Upload provider version index to S3
		err = index.UploadProviderVersionIndex(ctx, p.uploader, p.config.Bucket.BucketName, providerIndex)
		if err != nil {
			slog.WarnContext(ctx, "Failed to upload provider version index",
				"provider", fmt.Sprintf("%s/%s", namespace, name),
				"error", err)
		} else {
			slog.InfoContext(ctx, "Successfully uploaded provider version index",
				"provider", fmt.Sprintf("%s/%s", namespace, name),
				"versions", len(providerIndex.Versions))

			// Note: Global provider index is NOT updated here to avoid race conditions.
			// Use the `rebuild-global-indexes` command to rebuild it from the database.
		}
	}

	return responses, nil
}

// storeFailedVersion stores a minimal version record with status='failed' to prevent re-scraping
// Note: Repository and provider records are stored BEFORE parallel processing in IndexAllVersions
// provider parameter must be provided by the caller to avoid redundant file reads
func (p *ProviderReader) storeFailedVersion(ctx context.Context, namespace, name, version, errorMessage string, provider *registry.Provider) error {
	// Start database transaction
	tx, err := p.db.Begin(ctx)
	if err != nil {
		return fmt.Errorf("failed to start transaction: %w", err)
	}
	defer tx.Rollback(ctx)

	// Store provider version with status='failed' and error message
	err = storage.StoreProviderVersion(ctx, tx, namespace, name, version, 0, 0, nil, false, "failed", "processing_error", errorMessage)
	if err != nil {
		return fmt.Errorf("failed to store provider version: %w", err)
	}

	// Commit transaction
	err = tx.Commit(ctx)
	if err != nil {
		return fmt.Errorf("failed to commit transaction: %w", err)
	}

	slog.InfoContext(ctx, "Stored failed version record to prevent re-scraping",
		"provider", fmt.Sprintf("%s/%s", namespace, name),
		"version", version)

	return nil
}
