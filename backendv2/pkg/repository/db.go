package repository

import (
	"context"
	"fmt"

	"github.com/jackc/pgx/v5/pgxpool"
)

// StoreRepositoryStats stores repository statistics in the database
func StoreRepositoryStats(ctx context.Context, pool *pgxpool.Pool, metadata *RepositoryMetadata) error {
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
