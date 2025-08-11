package index

import (
	"context"
	"errors"
	"fmt"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/trace"

	"github.com/opentofu/registry-ui/pkg/config"
	"github.com/opentofu/registry-ui/pkg/docscraper"
	"github.com/opentofu/registry-ui/pkg/providers"
	"github.com/opentofu/registry-ui/pkg/registry"
)

// IndexService provides provider indexing functionality
type IndexService struct {
	pool    *pgxpool.Pool
	scraper *docscraper.Scraper
	config  *config.BackendConfig
	tracer  trace.Tracer
}

// NewIndexService creates a new IndexService with the required dependencies
func NewIndexService(cfg *config.BackendConfig, pool *pgxpool.Pool, scraper *docscraper.Scraper) *IndexService {
	return &IndexService{
		pool:    pool,
		scraper: scraper,
		config:  cfg,
		tracer:  otel.Tracer("opentofu-registry-backend"),
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
