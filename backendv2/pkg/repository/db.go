package repository

import (
	"context"
	"fmt"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"go.opentelemetry.io/otel/attribute"

	"github.com/opentofu/registry-ui/pkg/telemetry"
)

// ListRepositoriesForStatsSync returns repositories that have no stats datapoint
// newer than staleAfter ago. i.e. repos that are stale or have never been
// synced. A staleAfter of <= 0 returns every repository.
func ListRepositoriesForStatsSync(ctx context.Context, pool *pgxpool.Pool, staleAfter time.Duration) ([]RepoIdentifier, error) {
	ctx, span := telemetry.Tracer().Start(ctx, "repository.list_for_stats_sync")
	defer span.End()
	span.SetAttributes(attribute.Float64("stale_after_seconds", staleAfter.Seconds()))

	rows, err := pool.Query(ctx, `
		SELECT r.organisation, r.name
		FROM repositories r
		WHERE NOT EXISTS (
			SELECT 1 FROM repository_stats s
			WHERE s.repo_organisation = r.organisation
			  AND s.repo_name = r.name
			  AND s.recorded_at > now() - make_interval(secs => $1)
		)
		ORDER BY r.organisation, r.name`, staleAfter.Seconds())
	if err != nil {
		span.RecordError(err)
		return nil, fmt.Errorf("failed to query repositories: %w", err)
	}
	defer rows.Close()

	var repos []RepoIdentifier
	for rows.Next() {
		var r RepoIdentifier
		if err := rows.Scan(&r.Owner, &r.Name); err != nil {
			span.RecordError(err)
			return nil, fmt.Errorf("failed to scan repository row: %w", err)
		}
		repos = append(repos, r)
	}
	if err := rows.Err(); err != nil {
		span.RecordError(err)
		return nil, fmt.Errorf("error iterating repository rows: %w", err)
	}

	span.SetAttributes(attribute.Int("repositories.count", len(repos)))
	return repos, nil
}

// StoreRepositoryStatsBatch stores repository stats in the database
func StoreRepositoryStatsBatch(ctx context.Context, pool *pgxpool.Pool, stats map[RepoIdentifier]RepositoryStats) error {
	ctx, span := telemetry.Tracer().Start(ctx, "repository.store_stats_batch")
	defer span.End()
	span.SetAttributes(attribute.Int("repository.batch_size", len(stats)))

	if len(stats) == 0 {
		return nil
	}

	batch := &pgx.Batch{}
	for repo, s := range stats {
		batch.Queue(`
			INSERT INTO repository_stats (repo_organisation, repo_name, stars, forks, watchers, open_issues, subscribers, topics, recorded_at)
			VALUES ($1, $2, $3, $4, $5, $6, $7, $8, NOW())`,
			repo.Owner, repo.Name, s.Stars, s.Forks, s.Watchers, s.OpenIssues, s.Watchers, s.Topics)
	}

	br := pool.SendBatch(ctx, batch)
	defer br.Close()

	for range stats {
		if _, err := br.Exec(); err != nil {
			span.RecordError(err)
			return fmt.Errorf("failed to insert repository stats batch: %w", err)
		}
	}

	return nil
}

// StoreRepositoryStats stores repository statistics in the database
func StoreRepositoryStats(ctx context.Context, pool *pgxpool.Pool, metadata *RepositoryMetadata) error {

	ctx, span := telemetry.Tracer().Start(ctx, "repository.store_stats")
	defer span.End()
	query := `
		INSERT INTO repository_stats (repo_organisation, repo_name, stars, forks, watchers, open_issues, subscribers, topics, recorded_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, NOW())`

	// Handle potential redirect - use actual owner/name for storage
	owner := metadata.ActualOwner
	name := metadata.ActualName
	if owner == "" {
		owner = metadata.Owner
	}
	if name == "" {
		name = metadata.Name
	}

	_, err := pool.Exec(ctx, query,
		owner,
		name,
		metadata.Stars,
		metadata.Forks,
		metadata.Subscribers, // watchers field maps to subscribers
		metadata.OpenIssues,
		metadata.Subscribers,
		metadata.Topics,
	)
	if err != nil {
		return fmt.Errorf("failed to insert repository stats: %w", err)
	}

	return nil
}

// UpdateRepositoryMetadata updates repository metadata in the database
func UpdateRepositoryMetadata(ctx context.Context, pool *pgxpool.Pool, metadata *RepositoryMetadata) error {
	// Handle potential redirect - use actual owner/name for storage
	owner := metadata.ActualOwner
	name := metadata.ActualName
	if owner == "" {
		owner = metadata.Owner
	}
	if name == "" {
		name = metadata.Name
	}

	// Pass parent fields as nullable parameters — they will be nil for non-forks,
	// and EXCLUDED references handle both cases in a single query.
	var parentOwner, parentName *string
	if metadata.IsFork && metadata.ParentOwner != "" && metadata.ParentName != "" {
		parentOwner = &metadata.ParentOwner
		parentName = &metadata.ParentName
	}

	query := `
		INSERT INTO repositories (
			organisation, name, description, homepage, language, archived,
			default_branch, created_at_github, pushed_at, updated_at_github,
			is_fork, parent_organisation, parent_name
		)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13)
		ON CONFLICT (organisation, name) DO UPDATE SET
			description = EXCLUDED.description,
			homepage = EXCLUDED.homepage,
			language = EXCLUDED.language,
			archived = EXCLUDED.archived,
			default_branch = EXCLUDED.default_branch,
			created_at_github = EXCLUDED.created_at_github,
			pushed_at = EXCLUDED.pushed_at,
			updated_at_github = EXCLUDED.updated_at_github,
			is_fork = EXCLUDED.is_fork,
			parent_organisation = EXCLUDED.parent_organisation,
			parent_name = EXCLUDED.parent_name,
			updated_at = NOW()`

	_, err := pool.Exec(ctx, query,
		owner,
		name,
		metadata.Description,
		metadata.Homepage,
		metadata.Language,
		metadata.Archived,
		metadata.DefaultBranch,
		metadata.CreatedAtGitHub,
		metadata.PushedAt,
		metadata.UpdatedAtGitHub,
		metadata.IsFork,
		parentOwner,
		parentName,
	)
	if err != nil {
		return fmt.Errorf("failed to update repository metadata: %w", err)
	}

	return nil
}
