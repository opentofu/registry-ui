package index

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"log/slog"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/feature/s3/manager"
	"github.com/aws/aws-sdk-go-v2/service/s3"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/trace"

	"github.com/opentofu/registry-ui/pkg/config"
	"github.com/opentofu/registry-ui/pkg/docscraper"
	"github.com/opentofu/registry-ui/pkg/github"
	"github.com/opentofu/registry-ui/pkg/providers"
	"github.com/opentofu/registry-ui/pkg/registry"
)

// IndexService provides provider indexing functionality
type IndexService struct {
	pool         *pgxpool.Pool
	scraper      *docscraper.Scraper
	githubClient *github.Client
	s3Uploader   *manager.Uploader
	config       *config.BackendConfig
	tracer       trace.Tracer
}

// NewIndexService creates a new IndexService with the required dependencies
func NewIndexService(cfg *config.BackendConfig, pool *pgxpool.Pool, scraper *docscraper.Scraper, s3Client *s3.Client) *IndexService {
	return &IndexService{
		pool:         pool,
		scraper:      scraper,
		githubClient: github.NewClient(&cfg.GitHub),
		s3Uploader:   manager.NewUploader(s3Client),
		config:       cfg,
		tracer:       otel.Tracer("opentofu-registry-backend"),
	}
}

// prepareRegistry clones and updates the registry repository
func (s *IndexService) prepareRegistry(ctx context.Context) (*registry.Registry, error) {
	ctx, span := s.tracer.Start(ctx, "prepareRegistry")
	defer span.End()

	reg, err := registry.New(s.config.RegistryPath)
	if err != nil {
		span.RecordError(err)
		return nil, fmt.Errorf("failed to initialize registry: %w", err)
	}

	if err := reg.Clone(ctx); err != nil {
		span.RecordError(err)
		return nil, fmt.Errorf("failed to clone/update registry: %w", err)
	}

	if err := reg.Update(ctx); err != nil {
		span.RecordError(err)
		return nil, fmt.Errorf("failed to update registry: %w", err)
	}

	return reg, nil
}

// prepareProviderWithAlias sets up the provider for processing with alias support
func (s *IndexService) prepareProviderWithAlias(ctx context.Context, originalNamespace, originalName, canonicalNamespace, canonicalName string, isAlias bool) (*providers.Provider, error) {
	ctx, span := s.tracer.Start(ctx, "prepareProviderWithAlias")
	defer span.End()

	span.SetAttributes(
		attribute.String("provider.original_namespace", originalNamespace),
		attribute.String("provider.original_name", originalName),
		attribute.String("provider.canonical_namespace", canonicalNamespace),
		attribute.String("provider.canonical_name", canonicalName),
		attribute.Bool("provider.is_alias", isAlias),
	)

	// Store provider in database under original (requested) address
	if err := s.ensureProviderWithAliasInDB(ctx, originalNamespace, originalName, canonicalNamespace, canonicalName, isAlias); err != nil {
		span.RecordError(err)
		return nil, fmt.Errorf("failed to ensure provider in database: %w", err)
	}

	// Prepare repository using canonical address for documentation scraping
	_, err := s.prepareProviderRepo(ctx, canonicalNamespace, canonicalName)
	if err != nil {
		span.RecordError(err)
		return nil, fmt.Errorf("failed to prepare repository: %w", err)
	}

	// Create provider object using canonical address for repository operations
	return providers.NewProvider(s.config, canonicalNamespace, canonicalName), nil
}

// ResolveProviderAlias resolves provider aliases to their canonical addresses
func (s *IndexService) ResolveProviderAlias(ctx context.Context, namespace, name string) (canonicalNamespace, canonicalName string, isAlias bool, err error) {
	ctx, span := s.tracer.Start(ctx, "ResolveProviderAlias")
	defer span.End()

	// Query the provider_aliases table to check if this is an alias
	var targetNamespace, targetName string
	err = s.pool.QueryRow(ctx, `
		SELECT target_namespace, target_name 
		FROM provider_aliases 
		WHERE original_namespace = $1 AND original_name = $2
	`, namespace, name).Scan(&targetNamespace, &targetName)

	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			// Not an alias, return original namespace/name
			return namespace, name, false, nil
		}
		span.RecordError(err)
		return "", "", false, fmt.Errorf("failed to query provider aliases: %w", err)
	}

	// Found an alias, return the canonical address
	return targetNamespace, targetName, true, nil
}

// GenerateProviderIndex generates an index.json file for a provider
func (s *IndexService) GenerateProviderIndex(ctx context.Context, namespace, name string) error {
	ctx, span := s.tracer.Start(ctx, "index.GenerateProviderIndex")
	defer span.End()

	span.SetAttributes(
		attribute.String("provider.namespace", namespace),
		attribute.String("provider.name", name),
	)

	// Fetch GitHub metadata and update repository information
	repoName := fmt.Sprintf("terraform-provider-%s", name)
	repoMeta, err := s.githubClient.GetRepositoryMetadata(ctx, namespace, repoName)
	if err != nil {
		span.RecordError(err)
		return fmt.Errorf("failed to fetch GitHub metadata: %w", err)
	}

	// Query provider data from database
	providerData, err := s.queryProviderData(ctx, namespace, name)
	if err != nil {
		span.RecordError(err)
		return fmt.Errorf("failed to query provider data: %w", err)
	}

	// Update repository metadata in database
	if err := s.updateRepositoryMetadata(ctx, namespace, name, repoMeta); err != nil {
		span.RecordError(err)
		return fmt.Errorf("failed to update repository metadata: %w", err)
	}

	// Generate and upload index to S3
	if err := s.generateAndUploadIndex(ctx, namespace, name, providerData); err != nil {
		span.RecordError(err)
		return fmt.Errorf("failed to generate and upload index: %w", err)
	}

	return nil
}

// queryProviderData retrieves all provider data from the database
func (s *IndexService) queryProviderData(ctx context.Context, namespace, name string) (*DatabaseProviderData, error) {
	ctx, span := s.tracer.Start(ctx, "index.queryProviderData")
	defer span.End()

	span.SetAttributes(
		attribute.String("provider.namespace", namespace),
		attribute.String("provider.name", name),
	)

	query := `
		SELECT 
			p.namespace, p.name, p.warnings,
			r.organisation, r.name,
			r.stars, r.fork_count, r.description, r.is_fork, 
			r.parent_organisation, r.parent_name,
			ARRAY_AGG(pv.version ORDER BY pv.tag_created_at DESC) FILTER (WHERE pv.version IS NOT NULL) as versions,
			ARRAY_AGG(pv.tag_created_at ORDER BY pv.tag_created_at DESC) FILTER (WHERE pv.tag_created_at IS NOT NULL) as version_dates
		FROM providers p
		JOIN repositories r ON p.repo_organisation = r.organisation AND p.repo_name = r.name
		LEFT JOIN provider_versions pv ON p.namespace = pv.provider_namespace AND p.name = pv.provider_name
		WHERE p.namespace = $1 AND p.name = $2
		GROUP BY p.namespace, p.name, p.warnings, r.organisation, r.name, r.stars, r.fork_count, r.description, r.is_fork, r.parent_organisation, r.parent_name
	`

	var data DatabaseProviderData
	var parentOwner, parentName, description *string

	err := s.pool.QueryRow(ctx, query, namespace, name).Scan(
		&data.Namespace, &data.Name, &data.Warnings,
		&data.RepoOwner, &data.RepoName,
		&data.Stars, &data.ForkCount, &description, &data.IsFork,
		&parentOwner, &parentName,
		&data.Versions, &data.VersionDates,
	)

	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, fmt.Errorf("provider %s/%s not found in database", namespace, name)
		}
		return nil, fmt.Errorf("failed to query provider data: %w", err)
	}

	if description != nil {
		data.Description = *description
	}

	if parentOwner != nil && parentName != nil {
		data.ParentOwner = *parentOwner
		data.ParentName = *parentName
	}

	// Get reverse aliases
	reverseAliases, err := s.queryReverseAliases(ctx, namespace, name)
	if err != nil {
		return nil, fmt.Errorf("failed to query reverse aliases: %w", err)
	}
	data.ReverseAliases = reverseAliases

	// Get canonical address
	canonicalAddr, err := s.queryCanonicalAddr(ctx, namespace, name)
	if err != nil {
		return nil, fmt.Errorf("failed to query canonical address: %w", err)
	}
	data.CanonicalAddr = canonicalAddr

	// Set fork relationship
	if data.IsFork && data.ParentOwner != "" && data.ParentName != "" {
		data.ForkOf = &ProviderAddr{
			Display:   fmt.Sprintf("%s/%s", data.ParentOwner, data.ParentName),
			Namespace: data.ParentOwner,
			Name:      data.ParentName,
		}
	}

	span.SetAttributes(
		attribute.Int("versions.count", len(data.Versions)),
		attribute.Int64("repo.stars", data.Stars),
		attribute.Int64("repo.forks", data.ForkCount),
		attribute.Bool("repo.is_fork", data.IsFork),
	)

	return &data, nil
}

// queryReverseAliases gets aliases that point to this provider
func (s *IndexService) queryReverseAliases(ctx context.Context, targetNamespace, targetName string) ([]ProviderAddr, error) {
	query := `
		SELECT original_namespace, original_name 
		FROM provider_aliases 
		WHERE target_namespace = $1 AND target_name = $2
	`

	rows, err := s.pool.Query(ctx, query, targetNamespace, targetName)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var aliases []ProviderAddr
	for rows.Next() {
		var namespace, name string
		if err := rows.Scan(&namespace, &name); err != nil {
			continue
		}
		aliases = append(aliases, ProviderAddr{
			Display:   fmt.Sprintf("%s/%s", namespace, name),
			Namespace: namespace,
			Name:      name,
		})
	}

	return aliases, nil
}

// queryCanonicalAddr gets the canonical address if this is an alias
func (s *IndexService) queryCanonicalAddr(ctx context.Context, originalNamespace, originalName string) (*ProviderAddr, error) {
	query := `
		SELECT target_namespace, target_name 
		FROM provider_aliases 
		WHERE original_namespace = $1 AND original_name = $2
	`

	var targetNamespace, targetName string
	err := s.pool.QueryRow(ctx, query, originalNamespace, originalName).Scan(&targetNamespace, &targetName)
	if err != nil {
		if err == pgx.ErrNoRows {
			return nil, nil
		}
		return nil, err
	}

	return &ProviderAddr{
		Display:   fmt.Sprintf("%s/%s", targetNamespace, targetName),
		Namespace: targetNamespace,
		Name:      targetName,
	}, nil
}

// updateRepositoryMetadata updates repository metadata in the database
func (s *IndexService) updateRepositoryMetadata(ctx context.Context, namespace, name string, repoMeta *github.RepositoryMetadata) error {
	ctx, span := s.tracer.Start(ctx, "index.updateRepositoryMetadata")
	defer span.End()

	query := `
		UPDATE repositories 
		SET stars = $3, fork_count = $4, description = $5, is_fork = $6, 
		    parent_organisation = $7, parent_name = $8
		WHERE organisation = $1 AND name = $2
	`

	var parentOrg, parentName *string
	if repoMeta.IsFork {
		parentOrg = &repoMeta.ParentOwner
		parentName = &repoMeta.ParentName
	}

	_, err := s.pool.Exec(ctx, query,
		repoMeta.Owner, repoMeta.Name,
		repoMeta.Stars, repoMeta.Forks, repoMeta.Description, repoMeta.IsFork,
		parentOrg, parentName,
	)

	if err != nil {
		span.RecordError(err)
		return fmt.Errorf("failed to update repository metadata: %w", err)
	}

	span.SetAttributes(
		attribute.String("repo.owner", repoMeta.Owner),
		attribute.String("repo.name", repoMeta.Name),
		attribute.Int64("repo.stars", repoMeta.Stars),
		attribute.Int64("repo.forks", repoMeta.Forks),
		attribute.Bool("repo.is_fork", repoMeta.IsFork),
	)

	return nil
}

// generateAndUploadIndex creates and uploads the index.json file to S3
func (s *IndexService) generateAndUploadIndex(ctx context.Context, namespace, name string, data *DatabaseProviderData) error {
	ctx, span := s.tracer.Start(ctx, "index.generateAndUploadIndex")
	defer span.End()

	span.SetAttributes(
		attribute.String("provider.namespace", namespace),
		attribute.String("provider.name", name),
	)

	providerIndex := s.buildProviderIndex(data)

	jsonData, err := json.MarshalIndent(providerIndex, "", "  ")
	if err != nil {
		span.RecordError(err)
		return fmt.Errorf("failed to marshal provider index JSON: %w", err)
	}

	key := fmt.Sprintf("providers/%s/%s/index.json", namespace, name)

	slog.InfoContext(ctx, "Uploading provider index.json to S3",
		"provider", fmt.Sprintf("%s/%s", namespace, name),
		"s3_key", key,
		"json_size", len(jsonData))

	_, err = s.s3Uploader.Upload(ctx, &s3.PutObjectInput{
		Bucket:      aws.String(s.config.Bucket.BucketName),
		Key:         aws.String(key),
		Body:        bytes.NewReader(jsonData),
		ContentType: aws.String("application/json"),
	})

	if err != nil {
		span.RecordError(err)
		return fmt.Errorf("failed to upload index.json to S3: %w", err)
	}

	span.SetAttributes(
		attribute.String("s3.key", key),
		attribute.Int("json.size", len(jsonData)),
	)

	slog.InfoContext(ctx, "Successfully uploaded provider index.json",
		"provider", fmt.Sprintf("%s/%s", namespace, name),
		"s3_key", key)

	return nil
}

// buildProviderIndex constructs the ProviderIndexData from database data
func (s *IndexService) buildProviderIndex(data *DatabaseProviderData) *ProviderIndexData {
	versions := make([]ProviderVersionDescriptor, 0, len(data.Versions))

	for i, version := range data.Versions {
		descriptor := ProviderVersionDescriptor{
			ID: version,
		}
		if i < len(data.VersionDates) && !data.VersionDates[i].IsZero() {
			descriptor.Published = data.VersionDates[i]
		}
		versions = append(versions, descriptor)
	}

	link := fmt.Sprintf("https://github.com/%s/%s", data.RepoOwner, data.RepoName)

	index := &ProviderIndexData{
		Addr: ProviderAddr{
			Display:   fmt.Sprintf("%s/%s", data.Namespace, data.Name),
			Namespace: data.Namespace,
			Name:      data.Name,
		},
		Description:    data.Description,
		Stars:          data.Stars,
		ForkCount:      data.ForkCount,
		Versions:       versions,
		IsBlocked:      false,
		Link:           link,
		Warnings:       data.Warnings,
		ReverseAliases: data.ReverseAliases,
	}

	if data.CanonicalAddr != nil {
		index.CanonicalAddr = data.CanonicalAddr
	}

	if data.ForkOf != nil {
		index.ForkOf = data.ForkOf
	}

	return index
}
