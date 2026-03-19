package index

import (
	"context"
	"errors"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

// queryLatestRepositoryStats retrieves the latest repository statistics that we know about from the database
func queryLatestRepositoryStats(ctx context.Context, db *pgxpool.Pool, org, name string) (*RepositoryStats, error) {
	query := `
		SELECT stars, forks, watchers, subscribers, topics
		FROM repository_stats
		WHERE repo_organisation = $1 AND repo_name = $2
		ORDER BY recorded_at DESC
		LIMIT 1`

	var stats RepositoryStats
	err := db.QueryRow(ctx, query, org, name).Scan(
		&stats.Stars,
		&stats.Forks,
		&stats.Watchers,
		&stats.Subscribers,
		&stats.Topics,
	)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return &RepositoryStats{}, nil
		}
		return nil, err
	}

	return &stats, nil
}

// queryRepositoryMetadata retrieves repository metadata including fork information from the database
func queryRepositoryMetadata(ctx context.Context, db *pgxpool.Pool, org, name string) (*RepositoryMetadata, error) {
	query := `
		SELECT description, homepage, language, archived, default_branch,
		       is_fork, parent_organisation, parent_name
		FROM repositories
		WHERE organisation = $1 AND name = $2`

	var repo RepositoryMetadata
	err := db.QueryRow(ctx, query, org, name).Scan(
		&repo.Description,
		&repo.Homepage,
		&repo.Language,
		&repo.Archived,
		&repo.DefaultBranch,
		&repo.IsFork,
		&repo.ParentOrganisation,
		&repo.ParentName,
	)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return &RepositoryMetadata{}, nil
		}
		return nil, err
	}

	return &repo, nil
}

// queryModuleVersions retrieves all known versions for a module from the database
func queryModuleVersions(ctx context.Context, db *pgxpool.Pool, namespace, name, target string) ([]VersionInfo, error) {
	query := `
		SELECT version, discovered_at
		FROM module_versions
		WHERE module_namespace = $1 AND module_name = $2 AND module_target = $3
		ORDER BY safe_to_semver(version) DESC`

	rows, err := db.Query(ctx, query, namespace, name, target)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var versions []VersionInfo
	for rows.Next() {
		var version VersionInfo
		err := rows.Scan(&version.ID, &version.Published)
		if err != nil {
			return nil, err
		}
		versions = append(versions, version)
	}

	return versions, rows.Err()
}

// queryProviderVersions retrieves all known versions for a provider from the database
func queryProviderVersions(ctx context.Context, db *pgxpool.Pool, namespace, name string) ([]VersionInfo, error) {
	query := `
		SELECT version, discovered_at
		FROM provider_versions
		WHERE provider_namespace = $1 AND provider_name = $2
		ORDER BY safe_to_semver(version) DESC`

	rows, err := db.Query(ctx, query, namespace, name)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var versions []VersionInfo
	for rows.Next() {
		var version VersionInfo
		err := rows.Scan(&version.ID, &version.Published)
		if err != nil {
			return nil, err
		}
		versions = append(versions, version)
	}

	return versions, rows.Err()
}

// queryProviderWarnings retrieves known warnings for a provider  from the database
func queryProviderWarnings(ctx context.Context, db *pgxpool.Pool, namespace, name string) ([]string, error) {
	query := `
		SELECT warnings
		FROM providers
		WHERE namespace = $1 AND name = $2`

	var warnings []string
	err := db.QueryRow(ctx, query, namespace, name).Scan(&warnings)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return []string{}, nil
		}
		return nil, err
	}

	return warnings, nil
}
