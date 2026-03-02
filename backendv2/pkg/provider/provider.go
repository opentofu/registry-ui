package provider

import (
	"context"
	"fmt"
	"path/filepath"
	"strings"
	"time"

	"github.com/aws/aws-sdk-go-v2/feature/s3/manager"
	"github.com/aws/aws-sdk-go-v2/service/s3"
	"github.com/jackc/pgx/v5/pgxpool"
	"go.opentelemetry.io/otel/attribute"

	"github.com/opentofu/registry-ui/pkg/config"
	"github.com/opentofu/registry-ui/pkg/git"
	"github.com/opentofu/registry-ui/pkg/license"
	"github.com/opentofu/registry-ui/pkg/provider/storage"
	"github.com/opentofu/registry-ui/pkg/registry"
	"github.com/opentofu/registry-ui/pkg/repository"
	"github.com/opentofu/registry-ui/pkg/telemetry"
)

// ProviderReader is the main entry point for all provider operations
type ProviderReader struct {
	config       *config.BackendConfig
	db           *pgxpool.Pool
	s3Client     *s3.Client
	uploader     *manager.Uploader
	githubClient *repository.Client
}

// NewProviderReader creates a new ProviderReader with all dependencies initialized
func NewProviderReader(cfg *config.BackendConfig) (*ProviderReader, error) {
	// Initialize database pool
	db, err := cfg.DB.GetPool(context.Background())
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

	return &ProviderReader{
		config:       cfg,
		db:           db,
		s3Client:     s3Client,
		uploader:     uploader,
		githubClient: githubClient,
	}, nil
}

// EnsureParentRecords stores the repository and provider records in the database.
// Must be called before IndexVersion when IndexVersion is called standalone
// (IndexAllVersions handles this internally).
func (p *ProviderReader) EnsureParentRecords(ctx context.Context, provider *registry.Provider) error {
	namespace := provider.Namespace
	name := provider.Name
	repoOrg := namespace
	repoName := fmt.Sprintf("terraform-provider-%s", name)

	if err := storage.StoreRepository(ctx, p.db, repoOrg, repoName); err != nil {
		return fmt.Errorf("failed to store repository: %w", err)
	}

	if err := storage.StoreProvider(ctx, p.db, namespace, name, repoOrg, repoName, provider.Warnings); err != nil {
		return fmt.Errorf("failed to store provider: %w", err)
	}

	return nil
}

// ScrapeAllVersions scrapes all versions of a provider
// registryProvider must be provided by the caller
func (p *ProviderReader) ScrapeAllVersions(ctx context.Context, provider *registry.Provider) error {
	if provider == nil {
		panic("registryProvider cannot be nil, this is a developer error")
	}

	ctx, span := telemetry.Tracer().Start(ctx, "provider.scrape_all_versions")
	defer span.End()

	namespace := provider.Namespace
	name := provider.Name

	span.SetAttributes(
		attribute.String("provider.namespace", namespace),
		attribute.String("provider.name", name),
	)

	responses, err := p.IndexAllVersions(ctx, namespace, name, provider)
	if err != nil {
		return err
	}

	// Check if any versions failed
	var failedVersions []string
	for _, response := range responses {
		if !response.Success {
			failedVersions = append(failedVersions, response.Version)
		}
	}

	if len(failedVersions) > 0 {
		return fmt.Errorf("failed to scrape versions: %v", failedVersions)
	}

	return nil
}

// GetTagCreationDate returns the creation date of a git tag for a provider
func (p *ProviderReader) GetTagCreationDate(ctx context.Context, namespace, name, version string) (*time.Time, error) {
	repoURL := fmt.Sprintf("https://github.com/%s/terraform-provider-%s", namespace, name)
	localPath := filepath.Join(p.config.WorkDir, "providers", namespace, name)

	repo, err := git.GetRepo(repoURL, localPath)
	if err != nil {
		return nil, fmt.Errorf("failed to get repository: %w", err)
	}

	// Ensure version has 'v' prefix for provider repositories
	gitTag := version
	if !strings.HasPrefix(version, "v") {
		gitTag = "v" + version
	}

	return repo.GetTagDate(ctx, gitTag)
}

// CheckoutVersionForScraping creates a worktree for a tag and returns the directory path and cleanup function
// Provider tags typically have a 'v' prefix (e.g., v1.0.0), so we ensure the tag starts with 'v'
func (p *ProviderReader) CheckoutVersionForScraping(ctx context.Context, namespace, name, tag string) (string, func(), error) {
	repoURL := fmt.Sprintf("https://github.com/%s/terraform-provider-%s", namespace, name)
	localPath := filepath.Join(p.config.WorkDir, "providers", namespace, name)

	// Ensure tag has 'v' prefix for provider repositories
	gitTag := tag
	if !strings.HasPrefix(tag, "v") {
		gitTag = "v" + tag
	}

	return git.CheckoutVersionForScraping(ctx, repoURL, localPath, gitTag)
}

// DetectLicensesInDirectory detects licenses in a given directory path
func (p *ProviderReader) DetectLicensesInDirectory(ctx context.Context, namespace, name, directory string) (license.List, error) {
	ctx, span := telemetry.Tracer().Start(ctx, "provider.detect_licenses_in_directory")
	defer span.End()

	span.SetAttributes(
		attribute.String("provider.namespace", namespace),
		attribute.String("provider.name", name),
		attribute.String("directory", directory),
	)

	// Create license detector
	detector, err := license.New(p.config.License, p.githubClient)
	if err != nil {
		span.RecordError(err)
		return nil, fmt.Errorf("failed to create license detector: %w", err)
	}

	// Generate repo URL for context
	repoURL := fmt.Sprintf("https://github.com/%s/terraform-provider-%s", namespace, name)

	// Detect licenses in the directory
	licenses, err := detector.Detect(ctx, directory, repoURL)
	if err != nil {
		span.RecordError(err)
		return nil, fmt.Errorf("failed to detect licenses in directory %s: %w", directory, err)
	}

	span.SetAttributes(attribute.Int("provider.licenses_count", len(licenses)))
	return licenses, nil
}
