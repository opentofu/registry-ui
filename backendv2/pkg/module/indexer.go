package module

import (
	"context"
	"encoding/json"
	"fmt"
	"log/slog"
	"os"
	"os/exec"
	"path"
	"path/filepath"
	"slices"
	"strings"
	"sync"
	"time"

	"github.com/jackc/pgx/v5"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/codes"
	"golang.org/x/sync/errgroup"

	"github.com/opentofu/registry-ui/pkg/git"
	"github.com/opentofu/registry-ui/pkg/index"
	"github.com/opentofu/registry-ui/pkg/license"
	"github.com/opentofu/registry-ui/pkg/module/storage"
	"github.com/opentofu/registry-ui/pkg/registry"
	"github.com/opentofu/registry-ui/pkg/repository"
	"github.com/opentofu/registry-ui/pkg/telemetry"
	"github.com/opentofu/registry-ui/pkg/tofu"
)

// IndexVersion indexes a specific version of a module
// registryModule parameter must be provided by the caller to avoid redundant file reads
// This will NOT check if the version already exists - caller must ensure this
func (r *Reader) IndexVersion(ctx context.Context, namespace, name, target, version string, registryModule *registry.Module) (response *IndexResponse, err error) {
	ctx, span := telemetry.Tracer().Start(ctx, "module.index_version")
	defer func() {
		if err != nil {
			span.SetStatus(codes.Error, err.Error())
		}
		span.End()
	}()

	span.SetAttributes(
		attribute.String("registryModule.namespace", namespace),
		attribute.String("registryModule.name", name),
		attribute.String("registryModule.target", target),
		attribute.String("registryModule.version", version),
	)

	slog.DebugContext(ctx, "Starting registryModule version indexing",
		"registryModule", fmt.Sprintf("%s/%s/%s", namespace, name, target), "version", version)

	// Get the registryVersions from the registryModule
	registryVersions := registryModule.Versions
	versionExists := slices.Contains(registryVersions, version)

	if !versionExists {
		return nil, fmt.Errorf("version %s not found for module in the registry: %s/%s/%s", version, namespace, name, target)
	}

	// Checkout the version for processing
	var workDir string
	var cleanup func()
	workDir, cleanup, err = r.CheckoutVersionForScraping(ctx, namespace, name, target, version)
	if err != nil {
		span.RecordError(err)
		span.SetStatus(codes.Error, err.Error())
		return nil, fmt.Errorf("failed to checkout version: %w", err)
	}
	defer cleanup()

	// Detect licenses early to validate compatibility before expensive operations
	var licenses license.List
	licenses, err = r.DetectLicensesInDirectory(ctx, namespace, name, target, workDir)
	if err != nil {
		span.RecordError(err)
		span.SetStatus(codes.Error, err.Error())
		return nil, fmt.Errorf("failed to detect licenses: %w", err)
	}

	// Determine if version should be skipped due to license issues
	var shouldSkip bool
	var skipReason string

	// Handle no license found - mark for skip
	if len(licenses) == 0 {
		shouldSkip = true
		skipReason = "no_license"
		slog.WarnContext(ctx, "Module version has no license, will store with skipped status",
			"module", fmt.Sprintf("%s/%s/%s", namespace, name, target),
			"version", version)
		span.SetAttributes(
			attribute.String("module.skip_reason", "no_license"),
			attribute.Bool("module.version_skipped", true),
		)
	} else if licenses.HasIncompatible() {
		// Validate licenses - mark for skip if incompatible licenses found
		shouldSkip = true
		skipReason = "incompatible_license"
		incompatibleList := licenses.String()
		slog.WarnContext(ctx, "Module has incompatible license(s), will store with skipped status",
			"module", fmt.Sprintf("%s/%s/%s", namespace, name, target),
			"version", version,
			"licenses", incompatibleList)
		span.SetAttributes(
			attribute.String("module.skip_reason", "incompatible_license"),
			attribute.Bool("module.version_skipped", true),
			attribute.String("module.incompatible_licenses", incompatibleList),
		)
	}

	// Get tag creation date early - needed for both module data and database storage
	var tagCreatedAt *time.Time
	tagCreatedAt, err = r.GetTagCreationDate(ctx, namespace, name, target, version)
	if err != nil {
		// Log warning but don't fail - tag date is optional
		slog.WarnContext(ctx, "Failed to get tag creation date",
			"registryModule", fmt.Sprintf("%s/%s/%s", namespace, name, target),
			"version", version,
			"error", err)
		tagCreatedAt = nil
	}

	// Only process module data if license is acceptable
	var collectedData *CollectedModuleData
	var moduleData *ModuleData
	var indexChecksum, readmeChecksum string
	if !shouldSkip {
		// Build complete registryModule structure using parser (collects root, submodules, examples in parallel)
		collectedData, err = r.buildCompleteModuleData(ctx, namespace, name, target, version, workDir, licenses, tagCreatedAt)
		if err != nil {
			span.RecordError(err)
			span.SetStatus(codes.Error, err.Error())
			return nil, fmt.Errorf("failed to build complete registryModule data: %w", err)
		}
		moduleData = &collectedData.ModuleData

		// Store registryModule data in S3 and capture checksum
		indexChecksum, err = storage.StoreModuleInS3(ctx, r.uploader, r.config.Bucket.BucketName, namespace, name, target, version, moduleData)
		if err != nil {
			span.RecordError(err)
			span.SetStatus(codes.Error, err.Error())
			return nil, fmt.Errorf("failed to store registryModule in S3: %w", err)
		}

		// Store registryModule README in S3 and capture checksum
		readmeChecksum, err = storage.StoreModuleREADME(ctx, r.uploader, r.config.Bucket.BucketName, namespace, name, target, version, workDir)
		if err != nil {
			span.RecordError(err)
			span.SetStatus(codes.Error, err.Error())
			return nil, fmt.Errorf("failed to store registryModule README in S3: %w", err)
		}
	} else {
		// Create minimal empty module data for skipped versions
		moduleData = &ModuleData{}
	}

	// Start a database transaction for atomic operations
	// Note: Repository and module records are stored BEFORE parallel processing
	// in IndexAllVersions to avoid row lock contention between concurrent versions
	var tx pgx.Tx
	tx, err = r.db.Begin(ctx)
	if err != nil {
		span.RecordError(err)
		span.SetStatus(codes.Error, err.Error())
		return nil, fmt.Errorf("failed to start transaction: %w", err)
	}
	defer tx.Rollback(ctx)

	// Determine scrape status for database
	var scrapeStatus string
	if shouldSkip {
		scrapeStatus = "skipped"
	} else {
		scrapeStatus = "completed"
	}

	// Ensure repository and module records exist (these may already exist
	// when called from IndexAllVersions, but need to be created when
	// IndexVersion is called standalone)
	repoOrganisation := namespace
	repoName := fmt.Sprintf("terraform-%s-%s", target, name)

	if err := storage.StoreRepository(ctx, r.db, repoOrganisation, repoName); err != nil {
		span.RecordError(err)
		span.SetStatus(codes.Error, err.Error())
		return nil, fmt.Errorf("failed to store repository: %w", err)
	}

	if err := storage.StoreModule(ctx, r.db, namespace, name, target, repoOrganisation, repoName); err != nil {
		span.RecordError(err)
		span.SetStatus(codes.Error, err.Error())
		return nil, fmt.Errorf("failed to store module: %w", err)
	}

	// Store registryModule version in database (always store, even if skipped)
	err = storage.StoreModuleVersion(ctx, tx, namespace, name, target, version, moduleData, tagCreatedAt, scrapeStatus, skipReason, "", indexChecksum, readmeChecksum)
	if err != nil {
		span.RecordError(err)
		span.SetStatus(codes.Error, err.Error())
		return nil, fmt.Errorf("failed to store registryModule version: %w", err)
	}

	// Only store submodules and examples if not skipped
	if !shouldSkip {
		// Store submodules (data was already collected in buildCompleteModuleData)
		err = r.storeSubmodulesWithTx(ctx, tx, namespace, name, target, version, workDir, collectedData.Submodules)
		if err != nil {
			span.RecordError(err)
			span.SetStatus(codes.Error, err.Error())
			return nil, fmt.Errorf("failed to store submodules: %w", err)
		}

		// Store examples (data was already collected in buildCompleteModuleData)
		err = r.storeExamplesWithTx(ctx, tx, namespace, name, target, version, workDir, collectedData.Examples)
		if err != nil {
			span.RecordError(err)
			span.SetStatus(codes.Error, err.Error())
			return nil, fmt.Errorf("failed to store examples: %w", err)
		}
	}

	// Store license information regardless of skip status
	if len(licenses) > 0 {
		err = storage.StoreModuleVersionLicenses(ctx, tx, namespace, name, target, version, licenses)
		if err != nil {
			span.RecordError(err)
			span.SetStatus(codes.Error, err.Error())
			return nil, fmt.Errorf("failed to store module licenses: %w", err)
		}
	}

	// Commit transaction
	err = tx.Commit(ctx)
	if err != nil {
		span.RecordError(err)
		span.SetStatus(codes.Error, err.Error())
		return nil, fmt.Errorf("failed to commit transaction: %w", err)
	}

	// Count licenses from registryModule data
	licensesCount := len(moduleData.Licenses)

	response = &IndexResponse{
		Namespace:      namespace,
		Name:           name,
		Target:         target,
		Version:        version,
		ProcessedAt:    time.Now(),
		DocumentsCount: 1, // For modules, we typically store one main document
		LicensesCount:  licensesCount,
		Success:        true,
	}

	slog.DebugContext(ctx, "Successfully indexed module version",
		"registryModule", fmt.Sprintf("%s/%s/%s", namespace, name, target),
		"version", version)

	return response, nil
}

// IndexAllVersions indexes all versions of a module
// registryModule must be provided by the caller
func (r *Reader) IndexAllVersions(ctx context.Context, registryModule *registry.Module) ([]*IndexResponse, error) {
	if registryModule == nil {
		panic("registryModule cannot be nil")
	}

	namespace := registryModule.Namespace
	name := registryModule.Name
	target := registryModule.Target

	ctx, span := telemetry.Tracer().Start(ctx, "module.index_all_versions")
	defer span.End()

	span.SetAttributes(
		attribute.String("module.namespace", namespace),
		attribute.String("module.name", name),
		attribute.String("module.target", target),
	)

	// Get the versions from the module
	allVersions := registryModule.Versions

	// Get existing versions from database to avoid redundant work
	existingVersions, err := storage.GetExistingModuleVersions(ctx, r.db, namespace, name, target)
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
		attribute.Int("module.versions_total", len(allVersions)),
		attribute.Int("module.versions_existing", len(skippedVersions)),
		attribute.Int("module.versions_to_process", len(versionsToProcess)),
	)

	slog.DebugContext(ctx, "Version processing analysis",
		"module", fmt.Sprintf("%s/%s/%s", namespace, name, target),
		"total_versions", len(allVersions),
		"existing_versions", len(skippedVersions),
		"to_process", len(versionsToProcess))

	var responses []*IndexResponse
	var responsesMu sync.Mutex
	var failedVersions []string
	var policySkippedVersions []string // Versions skipped due to license policy
	var failedMu sync.Mutex

	// Add responses for skipped versions (already exist)
	for _, version := range skippedVersions {
		responses = append(responses, &IndexResponse{
			Namespace:      namespace,
			Name:           name,
			Target:         target,
			Version:        version,
			ProcessedAt:    time.Now(), // We don't query for exact time to keep it fast
			DocumentsCount: 0,          // We don't track counts for existing versions
			LicensesCount:  0,          // We don't track counts for existing versions
			Success:        true,
		})
	}

	// If no versions need processing, return early with skipped versions
	if len(versionsToProcess) == 0 {
		slog.DebugContext(ctx, "No versions to process, all versions already exist",
			"module", fmt.Sprintf("%s/%s/%s", namespace, name, target))
		return responses, nil
	}

	// Pre-clone and fetch tags before parallel processing to avoid race conditions
	repoURL := fmt.Sprintf("https://github.com/%s/terraform-%s-%s", namespace, target, name)
	localPath := fmt.Sprintf("%s/modules/%s/%s/%s", r.config.WorkDir, namespace, target, name)

	slog.DebugContext(ctx, "Preparing module repository for parallel processing",
		"module", fmt.Sprintf("%s/%s/%s", namespace, name, target),
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

	slog.DebugContext(ctx, "Successfully prepared repository for parallel processing",
		"module", fmt.Sprintf("%s/%s/%s", namespace, name, target))

	// Store repository and module ONCE before parallel version processing
	// This eliminates row lock contention between concurrent versions
	repoOrganisation := namespace
	repoName := fmt.Sprintf("terraform-%s-%s", target, name)

	if err := storage.StoreRepository(ctx, r.db, repoOrganisation, repoName); err != nil {
		span.RecordError(err)
		span.SetStatus(codes.Error, err.Error())
		return nil, fmt.Errorf("failed to store repository: %w", err)
	}

	if err := storage.StoreModule(ctx, r.db, namespace, name, target, repoOrganisation, repoName); err != nil {
		span.RecordError(err)
		span.SetStatus(codes.Error, err.Error())
		return nil, fmt.Errorf("failed to store module: %w", err)
	}

	slog.DebugContext(ctx, "Stored repository and module metadata",
		"module", fmt.Sprintf("%s/%s/%s", namespace, name, target))

	// Process each version that needs processing in parallel using errgroup
	g, gctx := errgroup.WithContext(ctx)
	g.SetLimit(r.config.Concurrency.Version)

	slog.DebugContext(ctx, "Starting parallel version processing",
		"module", fmt.Sprintf("%s/%s/%s", namespace, name, target),
		"versions_to_process", len(versionsToProcess),
		"concurrency_limit", r.config.Concurrency.Version)

	for _, version := range versionsToProcess {
		g.Go(func() error {
			// Create a new span for this version processing
			versionCtx, versionSpan := telemetry.Tracer().Start(gctx, "module.index_all_versions.version")
			defer versionSpan.End()

			versionSpan.SetAttributes(
				attribute.String("module.namespace", namespace),
				attribute.String("module.name", name),
				attribute.String("module.target", target),
				attribute.String("module.version", version),
			)

			// Process the version with its own context and span
			// Pass the module data to avoid redundant file reads
			response, err := r.IndexVersion(versionCtx, namespace, name, target, version, registryModule)
			if err != nil {
				// Check if this is a policy skip (license issue) or an actual failure
				errMsg := err.Error()
				isSkip := strings.Contains(errMsg, "module version skipped")

				if isSkip {
					// Policy-based skip (no license or incompatible license)
					// These are now stored in the database by IndexVersion with status='skipped'
					slog.WarnContext(versionCtx, "Skipped version due to license policy",
						"module", fmt.Sprintf("%s/%s/%s", namespace, name, target),
						"version", version, "reason", errMsg)
					versionSpan.SetAttributes(attribute.Bool("module.version_skipped", true))

					// Thread-safe append to skipped versions
					failedMu.Lock()
					policySkippedVersions = append(policySkippedVersions, version)
					failedMu.Unlock()
				} else {
					// Actual technical failure - store in database with status='failed' to prevent retry
					versionSpan.RecordError(err)
					versionSpan.SetStatus(codes.Error, err.Error())
					slog.WarnContext(versionCtx, "Failed to index version, storing as failed to prevent retry",
						"module", fmt.Sprintf("%s/%s/%s", namespace, name, target),
						"version", version, "error", err)

					// Store failed version in database to prevent re-scraping
					// Pass the module data to avoid redundant file reads
					storeErr := r.storeFailedVersion(versionCtx, namespace, name, target, version, err.Error(), registryModule)
					if storeErr != nil {
						slog.ErrorContext(versionCtx, "Failed to store failed version record",
							"module", fmt.Sprintf("%s/%s/%s", namespace, name, target),
							"version", version, "error", storeErr)
					}

					// Thread-safe append to failed versions
					failedMu.Lock()
					failedVersions = append(failedVersions, version)
					failedMu.Unlock()
				}

				// Create error response
				response = &IndexResponse{
					Namespace:    namespace,
					Name:         name,
					Target:       target,
					Version:      version,
					ProcessedAt:  time.Now(),
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

	// Log skipped versions (policy-based)
	if len(policySkippedVersions) > 0 {
		slog.WarnContext(ctx, "Some versions skipped due to license policy",
			"module", fmt.Sprintf("%s/%s/%s", namespace, name, target),
			"skipped_versions", policySkippedVersions)
	}

	// Log failed versions (technical failures)
	if len(failedVersions) > 0 {
		slog.WarnContext(ctx, "Some versions failed to index",
			"module", fmt.Sprintf("%s/%s/%s", namespace, name, target),
			"failed_versions", failedVersions)
	}

	// Sync repository metadata from GitHub (stats, fork info, etc.)
	if r.githubClient != nil {
		// Extract repository information from module source
		repoOrganisation, repoName := parseGitHubURL(registryModule.Source)
		if repoOrganisation != "" && repoName != "" {
			slog.DebugContext(ctx, "Syncing repository metadata from GitHub",
				"module", fmt.Sprintf("%s/%s/%s", namespace, name, target),
				"repository", fmt.Sprintf("%s/%s", repoOrganisation, repoName))

			err := repository.SyncRepositoryMetadata(ctx, r.db, r.githubClient, repoOrganisation, repoName)
			if err != nil {
				slog.ErrorContext(ctx, "Failed to sync repository metadata",
					"error", err,
					"module", fmt.Sprintf("%s/%s/%s", namespace, name, target),
					"repository", fmt.Sprintf("%s/%s", repoOrganisation, repoName))
				// Don't fail the sync process, just log the error
			}
		}
	}

	slog.DebugContext(ctx, "Generating module version index",
		"module", fmt.Sprintf("%s/%s/%s", namespace, name, target))

	moduleIndex, err := index.GenerateModuleVersionIndex(ctx, r.db, namespace, name, target)
	if err != nil {
		slog.WarnContext(ctx, "Failed to generate module version index",
			"module", fmt.Sprintf("%s/%s/%s", namespace, name, target),
			"error", err)
	} else {
		// Upload module version index to S3
		err = index.UploadModuleVersionIndex(ctx, r.uploader, r.config.Bucket.BucketName, moduleIndex)
		if err != nil {
			slog.WarnContext(ctx, "Failed to upload module version index",
				"module", fmt.Sprintf("%s/%s/%s", namespace, name, target),
				"error", err)
		} else {
			slog.DebugContext(ctx, "Successfully uploaded module version index",
				"module", fmt.Sprintf("%s/%s/%s", namespace, name, target),
				"versions", len(moduleIndex.Versions))

			// Note: Global module index is NOT updated here to avoid race conditions.
			// Use the `rebuild-global-indexes` command to rebuild it from the database.
		}
	}

	return responses, nil
}

// CollectedModuleData holds all collected module data for storage
type CollectedModuleData struct {
	ModuleData ModuleData
	Submodules map[string]SubmoduleData
	Examples   map[string]ExampleData
}

// buildCompleteModuleData builds the complete module structure by collecting and parsing all module data in parallel
func (r *Reader) buildCompleteModuleData(ctx context.Context, namespace, name, target, version, workDir string, licenses license.List, publishedAt *time.Time) (*CollectedModuleData, error) {
	ctx, span := telemetry.Tracer().Start(ctx, "module.build_complete_module_data")
	defer span.End()

	parser := NewModuleParser(workDir, namespace, name, target, version, publishedAt)

	// Collect root module, submodules, and examples in parallel
	var rootModuleData *tofu.Config
	var submodules map[string]SubmoduleData
	var examples map[string]ExampleData
	var rootErr, submodulesErr, examplesErr error

	g, gctx := errgroup.WithContext(ctx)

	g.Go(func() error {
		rootModuleData, _, rootErr = r.runTofuShow(gctx, workDir)
		return rootErr
	})

	g.Go(func() error {
		submodules, submodulesErr = r.collectSubmodulesDataParallel(gctx, namespace, name, target, version, workDir)
		return submodulesErr
	})

	g.Go(func() error {
		examples, examplesErr = r.collectExamplesDataParallel(gctx, namespace, name, target, version, workDir)
		return examplesErr
	})

	if err := g.Wait(); err != nil {
		span.RecordError(err)
		span.SetStatus(codes.Error, err.Error())
		return nil, fmt.Errorf("failed to collect module data: %w", err)
	}

	completeStructure, err := parser.BuildCompleteModuleStructure(ctx, workDir, rootModuleData, submodules, examples, licenses)
	if err != nil {
		span.RecordError(err)
		span.SetStatus(codes.Error, err.Error())
		return nil, fmt.Errorf("failed to build complete module structure: %w", err)
	}

	return &CollectedModuleData{
		ModuleData: completeStructure,
		Submodules: submodules,
		Examples:   examples,
	}, nil
}

// runTofuShow executes tofu show -json -module=DIR and returns the parsed JSON
// Returns (config, stderr, error) - stderr is returned for callers to store in SchemaError if needed
func (r *Reader) runTofuShow(ctx context.Context, moduleDir string) (*tofu.Config, string, error) {
	ctx, span := telemetry.Tracer().Start(ctx, "module.run_tofu_show")
	defer span.End()

	span.SetAttributes(
		attribute.String("module.dir", moduleDir),
	)

	cwd, err := os.Getwd()
	if err != nil {
		span.RecordError(err)
		span.SetStatus(codes.Error, err.Error())
		return nil, "", fmt.Errorf("failed to get current working directory: %w", err)
	}

	// Execute `tofu show -json -module=moduleDir`
	cmd := exec.CommandContext(ctx, path.Join(cwd, "tofu"), "show", "-json", "-module="+moduleDir)
	cmd.Dir = moduleDir

	slog.DebugContext(ctx, "Executing tofu show", "cmd", cmd.String(), "dir", moduleDir)

	var stdout, stderr strings.Builder
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr

	err = cmd.Run()

	// Return stderr to caller for storage in SchemaError (don't log it to avoid large traces)
	stderrStr := stderr.String()

	if err != nil {
		span.RecordError(err)
		span.SetStatus(codes.Error, err.Error())
		return nil, stderrStr, fmt.Errorf("tofu show failed: %w", err)
	}

	output := stdout.String()
	span.SetAttributes(
		attribute.Int("tofu.output_size", len(output)),
	)

	// Parse the JSON output
	var moduleData *tofu.Config
	if err := json.Unmarshal([]byte(output), &moduleData); err != nil {
		span.RecordError(err)
		span.SetStatus(codes.Error, err.Error())
		return nil, stderrStr, fmt.Errorf("failed to parse tofu JSON output: %w", err)
	}

	slog.DebugContext(ctx, "Successfully executed tofu show",
		"dir", moduleDir,
		"outputSize", len(output))

	return moduleData, stderrStr, nil
}

// collectSubmodulesDataParallel collects submodule data in parallel using tofu show.
// Returns a map of submodule name to SubmoduleData. This runs tofu show only once per submodule.
func (r *Reader) collectSubmodulesDataParallel(ctx context.Context, namespace, name, target, version, workDir string) (map[string]SubmoduleData, error) {
	ctx, span := telemetry.Tracer().Start(ctx, "module.collect_submodules_data")
	defer span.End()

	modulesDir := filepath.Join(workDir, "modules")

	// Check if modules directory exists
	if _, err := os.Stat(modulesDir); os.IsNotExist(err) {
		return make(map[string]SubmoduleData), nil // No submodules
	}

	// Read the modules directory
	entries, err := os.ReadDir(modulesDir)
	if err != nil {
		span.RecordError(err)
		span.SetStatus(codes.Error, err.Error())
		return nil, fmt.Errorf("failed to read modules directory: %w", err)
	}

	// Collect submodules concurrently using errgroup
	g, gctx := errgroup.WithContext(ctx)
	g.SetLimit(r.config.Concurrency.Submodule)

	var mu sync.Mutex
	submodules := make(map[string]SubmoduleData)

	slog.DebugContext(ctx, "Starting parallel submodule collection",
		"module", fmt.Sprintf("%s/%s/%s", namespace, name, target),
		"version", version,
		"submodules_count", len(entries),
		"concurrency_limit", r.config.Concurrency.Submodule)

	for _, entry := range entries {
		if !entry.IsDir() {
			continue
		}

		g.Go(func() error {
			submoduleName := entry.Name()
			fullSubmodulePath := filepath.Join(workDir, "modules", submoduleName)

			// Run tofu show on the submodule
			tofuConfig, _, err := r.runTofuShow(gctx, fullSubmodulePath)
			if err != nil {
				slog.WarnContext(gctx, "Failed to run tofu show on submodule",
					"submodule", submoduleName, "error", err)
				return nil // Don't fail the entire process for one submodule
			}

			// Create parser and transform the raw tofu config
			parser := NewModuleParser(workDir, namespace, name, target, version, nil)
			submoduleData, err := parser.BuildSubmoduleData(gctx, submoduleName, tofuConfig)
			if err != nil {
				slog.WarnContext(gctx, "Failed to transform submodule data",
					"submodule", submoduleName, "error", err)
				return nil // Don't fail the entire process for one submodule
			}

			mu.Lock()
			submodules[submoduleName] = submoduleData
			mu.Unlock()

			return nil
		})
	}

	if err := g.Wait(); err != nil {
		span.RecordError(err)
		span.SetStatus(codes.Error, err.Error())
		return nil, err
	}

	slog.DebugContext(ctx, "Completed parallel submodule collection",
		"module", fmt.Sprintf("%s/%s/%s", namespace, name, target),
		"version", version,
		"collected_count", len(submodules))

	return submodules, nil
}

// collectExamplesDataParallel collects example data in parallel using tofu show.
// Returns a map of example name to ExampleData. This runs tofu show only once per example.
func (r *Reader) collectExamplesDataParallel(ctx context.Context, namespace, name, target, version, workDir string) (map[string]ExampleData, error) {
	ctx, span := telemetry.Tracer().Start(ctx, "module.collect_examples_data")
	defer span.End()

	examplesDir := filepath.Join(workDir, "examples")

	// Check if examples directory exists
	if _, err := os.Stat(examplesDir); os.IsNotExist(err) {
		return make(map[string]ExampleData), nil // No examples
	}

	// Read the examples directory
	entries, err := os.ReadDir(examplesDir)
	if err != nil {
		span.RecordError(err)
		span.SetStatus(codes.Error, err.Error())
		return nil, fmt.Errorf("failed to read examples directory: %w", err)
	}

	// Collect examples concurrently using errgroup
	g, gctx := errgroup.WithContext(ctx)
	g.SetLimit(r.config.Concurrency.Example)

	var mu sync.Mutex
	examples := make(map[string]ExampleData)

	slog.DebugContext(ctx, "Starting parallel example collection",
		"module", fmt.Sprintf("%s/%s/%s", namespace, name, target),
		"version", version,
		"examples_count", len(entries),
		"concurrency_limit", r.config.Concurrency.Example)

	for _, entry := range entries {
		if !entry.IsDir() {
			continue
		}

		g.Go(func() error {
			exampleName := entry.Name()
			fullExamplePath := filepath.Join(workDir, "examples", exampleName)

			// Run tofu show on the example
			tofuConfig, _, err := r.runTofuShow(gctx, fullExamplePath)
			if err != nil {
				slog.WarnContext(gctx, "Failed to run tofu show on example",
					"example", exampleName, "error", err)
				return nil // Don't fail the entire process for one example
			}

			// Create parser and transform the raw tofu config
			parser := NewModuleParser(workDir, namespace, name, target, version, nil)
			exampleData, err := parser.BuildExampleData(gctx, exampleName, tofuConfig)
			if err != nil {
				slog.WarnContext(gctx, "Failed to transform example data",
					"example", exampleName, "error", err)
				return nil // Don't fail the entire process for one example
			}

			mu.Lock()
			examples[exampleName] = exampleData
			mu.Unlock()

			return nil
		})
	}

	if err := g.Wait(); err != nil {
		span.RecordError(err)
		span.SetStatus(codes.Error, err.Error())
		return nil, err
	}

	slog.DebugContext(ctx, "Completed parallel example collection",
		"module", fmt.Sprintf("%s/%s/%s", namespace, name, target),
		"version", version,
		"collected_count", len(examples))

	return examples, nil
}

// storeSubmodulesWithTx stores pre-collected submodule data to S3 and database.
// Data should be collected first using collectSubmodulesDataParallel.
func (r *Reader) storeSubmodulesWithTx(ctx context.Context, tx pgx.Tx, namespace, name, target, version, workDir string, submodules map[string]SubmoduleData) error {
	ctx, span := telemetry.Tracer().Start(ctx, "module.store_submodules")
	defer span.End()

	if len(submodules) == 0 {
		return nil
	}

	// Store submodules concurrently using errgroup
	g, gctx := errgroup.WithContext(ctx)
	g.SetLimit(r.config.Concurrency.Submodule)

	var txMu sync.Mutex

	slog.DebugContext(ctx, "Starting parallel submodule storage",
		"module", fmt.Sprintf("%s/%s/%s", namespace, name, target),
		"version", version,
		"submodules_count", len(submodules),
		"concurrency_limit", r.config.Concurrency.Submodule)

	for submoduleName, submoduleData := range submodules {
		g.Go(func() error {
			submodulePath := filepath.Join("modules", submoduleName)

			// Store submodule data in S3 and capture checksums
			indexChecksum, readmeChecksum, err := storage.StoreModuleSubmoduleInS3(gctx, r.uploader, r.config.Bucket.BucketName, namespace, name, target, version, submoduleName, submoduleData, workDir)
			if err != nil {
				slog.ErrorContext(gctx, "Failed to store submodule in S3",
					"submodule", submoduleName, "error", err)
				return fmt.Errorf("failed to store submodule %s in S3: %w", submoduleName, err)
			}

			txMu.Lock()
			err = storage.StoreModuleSubmodule(gctx, tx, namespace, name, target, version, submoduleName, submodulePath, submoduleData, indexChecksum, readmeChecksum)
			txMu.Unlock()
			if err != nil {
				slog.ErrorContext(gctx, "Failed to store submodule in database",
					"submodule", submoduleName, "error", err)
				return fmt.Errorf("failed to store submodule %s: %w", submoduleName, err)
			}

			slog.DebugContext(gctx, "Successfully stored submodule",
				"module", fmt.Sprintf("%s/%s/%s", namespace, name, target),
				"version", version,
				"submodule", submoduleName)

			return nil
		})
	}

	// Wait for all submodules to complete
	if err := g.Wait(); err != nil {
		span.RecordError(err)
		span.SetStatus(codes.Error, err.Error())
		return err
	}

	return nil
}

// storeExamplesWithTx stores pre-collected example data to S3 and database.
// Data should be collected first using collectExamplesDataParallel.
func (r *Reader) storeExamplesWithTx(ctx context.Context, tx pgx.Tx, namespace, name, target, version, workDir string, examples map[string]ExampleData) error {
	ctx, span := telemetry.Tracer().Start(ctx, "module.store_examples")
	defer span.End()

	if len(examples) == 0 {
		return nil
	}

	// Store examples concurrently using errgroup
	g, gctx := errgroup.WithContext(ctx)
	g.SetLimit(r.config.Concurrency.Example)

	var txMu sync.Mutex

	slog.DebugContext(ctx, "Starting parallel example storage",
		"module", fmt.Sprintf("%s/%s/%s", namespace, name, target),
		"version", version,
		"examples_count", len(examples),
		"concurrency_limit", r.config.Concurrency.Example)

	for exampleName, exampleData := range examples {
		g.Go(func() error {
			examplePath := filepath.Join("examples", exampleName)

			// Store example data in S3 and capture checksums
			indexChecksum, readmeChecksum, err := storage.StoreModuleExampleInS3(gctx, r.uploader, r.config.Bucket.BucketName, namespace, name, target, version, exampleName, exampleData, workDir)
			if err != nil {
				slog.ErrorContext(gctx, "Failed to store example in S3",
					"example", exampleName, "error", err)
				return fmt.Errorf("failed to store example %s in S3: %w", exampleName, err)
			}

			txMu.Lock()
			err = storage.StoreModuleExample(gctx, tx, namespace, name, target, version, exampleName, examplePath, exampleData, indexChecksum, readmeChecksum)
			txMu.Unlock()
			if err != nil {
				slog.ErrorContext(gctx, "Failed to store example in database",
					"example", exampleName, "error", err)
				return fmt.Errorf("failed to store example %s: %w", exampleName, err)
			}

			slog.DebugContext(gctx, "Successfully stored example",
				"module", fmt.Sprintf("%s/%s/%s", namespace, name, target),
				"version", version,
				"example", exampleName)

			return nil
		})
	}

	// Wait for all examples to complete
	if err := g.Wait(); err != nil {
		span.RecordError(err)
		span.SetStatus(codes.Error, err.Error())
		return err
	}

	return nil
}

// parseGitHubURL extracts organization and repository name from a GitHub URL
// Expected format: https://github.com/{org}/terraform-{target}-module-{name}
func parseGitHubURL(gitHubURL string) (organisation, repoName string) {
	// Remove https://github.com/ prefix
	if after, ok := strings.CutPrefix(gitHubURL, "https://github.com/"); ok {
		prefix := after
		parts := strings.Split(prefix, "/")
		if len(parts) >= 2 {
			return parts[0], parts[1]
		}
	}

	// Fallback: return empty strings if URL format is unexpected
	return "", ""
}

// storeFailedVersion stores a minimal version record with status='failed' to prevent re-scraping
// Note: Repository and module records are stored BEFORE parallel processing in IndexAllVersions
// registryModule parameter must be provided by the caller to avoid redundant file reads
func (r *Reader) storeFailedVersion(ctx context.Context, namespace, name, target, version, errorMessage string, registryModule *registry.Module) error {
	// Start database transaction
	tx, err := r.db.Begin(ctx)
	if err != nil {
		return fmt.Errorf("failed to start transaction: %w", err)
	}
	defer tx.Rollback(ctx)

	// Store module version with status='failed' and error message (no checksums for failed versions)
	err = storage.StoreModuleVersion(ctx, tx, namespace, name, target, version, &ModuleData{}, nil, "failed", "processing_error", errorMessage, "", "")
	if err != nil {
		return fmt.Errorf("failed to store module version: %w", err)
	}

	// Commit transaction
	err = tx.Commit(ctx)
	if err != nil {
		return fmt.Errorf("failed to commit transaction: %w", err)
	}

	slog.InfoContext(ctx, "Stored failed version record to prevent re-scraping",
		"module", fmt.Sprintf("%s/%s/%s", namespace, name, target),
		"version", version)

	return nil
}
