package module

import (
	"context"
	"fmt"
	"log/slog"
	"os"
	"path"
	"strings"
	"time"

	"github.com/aws/aws-sdk-go-v2/feature/s3/manager"
	"github.com/aws/aws-sdk-go-v2/service/s3"
	"github.com/jackc/pgx/v5/pgxpool"
	"go.opentelemetry.io/otel/attribute"

	"github.com/opentofu/registry-ui/pkg/config"
	"github.com/opentofu/registry-ui/pkg/git"
	"github.com/opentofu/registry-ui/pkg/license"
	"github.com/opentofu/registry-ui/pkg/registry"
	"github.com/opentofu/registry-ui/pkg/repository"
	"github.com/opentofu/registry-ui/pkg/telemetry"
)

// Reader is the main entry point for all module operations
type Reader struct {
	config       *config.BackendConfig
	db           *pgxpool.Pool
	s3Client     *s3.Client
	uploader     *manager.Uploader
	githubClient *repository.Client
	tofuPath     string
}

// NewModuleReader creates a new Reader with all dependencies initialized
func NewModuleReader(ctx context.Context, cfg *config.BackendConfig) (*Reader, error) {
	// Initialize database pool
	db, err := cfg.DB.GetPool(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to initialize database pool: %w", err)
	}

	// Initialize S3 client
	s3Client, err := cfg.Bucket.GetClient(context.Background())
	if err != nil {
		return nil, fmt.Errorf("failed to initialize S3 client: %w", err)
	}

	// Initialize S3 uploader
	uploader := manager.NewUploader(s3Client)

	// Initialize GitHub client if configured
	var githubClient *repository.Client
	if cfg.GitHub.Token != "" {
		githubClient = repository.NewClient(&cfg.GitHub)
	}

	// Ensure the tofu binary exists
	cwd, err := os.Getwd()
	if err != nil {
		return nil, fmt.Errorf("failed to get current working directory: %w", err)
	}

	tofuPath := path.Join(cwd, "tofu")
	stat, err := os.Stat(tofuPath)
	if err != nil || stat.IsDir() {
		return nil, fmt.Errorf("tofu binary not found in current directory: %w", err)
	}

	return &Reader{
		config:       cfg,
		db:           db,
		s3Client:     s3Client,
		uploader:     uploader,
		githubClient: githubClient,

		tofuPath: tofuPath,
	}, nil
}

// ScrapeVersion scrapes a specific version of a module
func (r *Reader) ScrapeVersion(ctx context.Context, module *registry.Module, version string) error {
	ctx, span := telemetry.Tracer().Start(ctx, "module.scrape_version")
	defer span.End()

	namespace := module.Namespace
	name := module.Name
	target := module.Target

	span.SetAttributes(
		attribute.String("module.namespace", namespace),
		attribute.String("module.name", name),
		attribute.String("module.target", target),
		attribute.String("module.version", version),
	)

	// Use the indexer to process the module version
	_, err := r.IndexVersion(ctx, namespace, name, target, version, module)
	return err
}

// ScrapeAllVersions scrapes all versions of a module
// registryModule can be provided to avoid redundant file reads (optional - will fetch if nil)
func (r *Reader) ScrapeAllVersions(ctx context.Context, module *registry.Module) error {
	ctx, span := telemetry.Tracer().Start(ctx, "module.scrape_all_versions")
	defer span.End()

	if module == nil {
		panic("registryModule must be provided, this is a developer error")
	}

	span.SetAttributes(
		attribute.String("module.namespace", module.Namespace),
		attribute.String("module.name", module.Name),
		attribute.String("module.target", module.Target),
	)

	// Use the indexer to process all versions
	responses, err := r.IndexAllVersions(ctx, module)
	if err != nil {
		return err
	}

	// Separate skipped versions from truly failed versions
	var failedVersions []string
	var skippedVersions []string
	for _, response := range responses {
		if !response.Success {
			// Check if this was a policy skip or an actual failure
			if strings.Contains(response.ErrorMessage, "skipped") {
				skippedVersions = append(skippedVersions, response.Version)
			} else {
				failedVersions = append(failedVersions, response.Version)
			}
		}
	}

	// Only return error if there were actual technical failures (not policy skips)
	if len(failedVersions) > 0 {
		return fmt.Errorf("failed to scrape versions: %v", failedVersions)
	}

	// Module sync is successful even if versions were skipped due to policy
	if len(skippedVersions) > 0 {
		slog.InfoContext(ctx, "Module sync completed with skipped versions",
			"module", fmt.Sprintf("%s/%s/%s", module.Namespace, module.Name, module.Target),
			"skipped", skippedVersions)
	}

	return nil
}

// GetTagCreationDate returns the creation date of a git tag for a module
func (r *Reader) GetTagCreationDate(ctx context.Context, namespace, name, target, version string) (*time.Time, error) {
	repoURL := fmt.Sprintf("https://github.com/%s/terraform-%s-%s", namespace, target, name)
	localPath := fmt.Sprintf("%s/modules/%s/%s/%s", r.config.WorkDir, namespace, target, name)

	repo, err := git.GetRepo(repoURL, localPath)
	if err != nil {
		return nil, fmt.Errorf("failed to get repository: %w", err)
	}

	return repo.GetTagDate(ctx, version)
}

// CheckoutVersionForScraping creates a worktree for a module tag and returns the directory path and cleanup function
func (r *Reader) CheckoutVersionForScraping(ctx context.Context, namespace, name, target, tag string) (string, func(), error) {
	repoURL := fmt.Sprintf("https://github.com/%s/terraform-%s-%s", namespace, target, name)
	localPath := fmt.Sprintf("%s/modules/%s/%s/%s", r.config.WorkDir, namespace, target, name)

	return git.CheckoutVersionForScraping(ctx, repoURL, localPath, tag)
}

// DetectLicensesInDirectory detects licenses in a given module directory path
func (r *Reader) DetectLicensesInDirectory(ctx context.Context, namespace, name, target, directory string) (license.List, error) {
	ctx, span := telemetry.Tracer().Start(ctx, "module.detect_licenses_in_directory")
	defer span.End()

	span.SetAttributes(
		attribute.String("module.namespace", namespace),
		attribute.String("module.name", name),
		attribute.String("module.target", target),
		attribute.String("directory", directory),
	)

	slog.DebugContext(ctx, "Starting license detection in module directory", "directory", directory)

	// Create license detector
	detector, err := license.New(r.config.License, r.githubClient)
	if err != nil {
		span.RecordError(err)
		return nil, fmt.Errorf("failed to create license detector: %w", err)
	}

	// Generate repo URL for context
	repoURL := fmt.Sprintf("https://github.com/%s/terraform-%s-%s", namespace, target, name)

	// Detect licenses in the directory
	licenses, err := detector.Detect(ctx, directory, repoURL)
	if err != nil {
		span.RecordError(err)
		return nil, fmt.Errorf("failed to detect licenses in directory %s: %w", directory, err)
	}

	span.SetAttributes(attribute.Int("module.licenses_count", len(licenses)))
	return licenses, nil
}
