package index

import (
	"context"
	"fmt"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/opentofu/registry-ui/pkg/telemetry"
)

// GenerateModuleVersionIndex creates a complete module version index from database data
func GenerateModuleVersionIndex(ctx context.Context, db *pgxpool.Pool, namespace, name, target string) (*ModuleVersionIndex, error) {
	repoOrg := namespace
	repoName := name

	ctx, span := telemetry.Tracer().Start(ctx, "index.generate_module_version")
	defer span.End()

	// Query repository stats
	stats, err := queryLatestRepositoryStats(ctx, db, repoOrg, repoName)
	if err != nil {
		return nil, fmt.Errorf("failed to query repository stats: %w", err)
	}

	// Query repository metadata (fork info)
	repo, err := queryRepositoryMetadata(ctx, db, repoOrg, repoName)
	if err != nil {
		return nil, fmt.Errorf("failed to query repository metadata: %w", err)
	}

	// Query all module versions
	versions, err := queryModuleVersions(ctx, db, namespace, name, target)
	if err != nil {
		return nil, fmt.Errorf("failed to query module versions: %w", err)
	}

	// Build index structure
	description := ""
	if repo.Description != nil {
		description = *repo.Description
	}

	index := &ModuleVersionIndex{
		Addr: ModuleAddr{
			Display:   fmt.Sprintf("%s/%s/%s", namespace, name, target),
			Namespace: namespace,
			Name:      name,
			Target:    target,
		},
		Description:        description,
		Versions:           versions,
		IsBlocked:          false,
		Popularity:         stats.Stars,
		ForkCount:          stats.Forks,
		UpstreamPopularity: 0, // Will be set below if this is a fork
		UpstreamForkCount:  0, // Will be set below if this is a fork
	}

	//  Add fork information if applicable
	if repo.IsFork && repo.ParentOrganisation != nil && repo.ParentName != nil {
		// Create fork_of address
		index.ForkOf = &ModuleAddr{
			Display:   fmt.Sprintf("%s/%s/%s", *repo.ParentOrganisation, *repo.ParentName, target),
			Namespace: *repo.ParentOrganisation,
			Name:      *repo.ParentName,
			Target:    target,
		}

		// Set fork_of_link to GitHub URL
		githubURL := fmt.Sprintf("https://github.com/%s/%s", *repo.ParentOrganisation, *repo.ParentName)
		index.ForkOfLink = &githubURL

		// Query upstream repository stats
		upstreamStats, err := queryLatestRepositoryStats(ctx, db, *repo.ParentOrganisation, *repo.ParentName)
		if err == nil {
			index.UpstreamPopularity = upstreamStats.Stars
			index.UpstreamForkCount = upstreamStats.Forks
		}
		// If upstream stats query fails, we keep the zero values
	}

	return index, nil
}

// GenerateProviderVersionIndex creates a complete provider version index from database data
func GenerateProviderVersionIndex(ctx context.Context, db *pgxpool.Pool, namespace, name string) (*ProviderVersionIndex, error) {
	ctx, span := telemetry.Tracer().Start(ctx, "index.generate_provider_version")
	defer span.End()

	repoOrg := namespace
	repoName := fmt.Sprintf("terraform-provider-%s", name)

	// Query repository stats
	stats, err := queryLatestRepositoryStats(ctx, db, repoOrg, repoName)
	if err != nil {
		return nil, fmt.Errorf("failed to query repository stats: %w", err)
	}

	// Query repository metadata (fork info)
	repo, err := queryRepositoryMetadata(ctx, db, repoOrg, repoName)
	if err != nil {
		return nil, fmt.Errorf("failed to query repository metadata: %w", err)
	}

	// Query all provider versions
	versions, err := queryProviderVersions(ctx, db, namespace, name)
	if err != nil {
		return nil, fmt.Errorf("failed to query provider versions: %w", err)
	}

	// Query provider warnings
	warnings, err := queryProviderWarnings(ctx, db, namespace, name)
	if err != nil {
		return nil, fmt.Errorf("failed to query provider warnings: %w", err)
	}

	// Build index structure
	description := ""
	if repo.Description != nil {
		description = *repo.Description
	}

	index := &ProviderVersionIndex{
		Addr: ProviderAddr{
			Display:   fmt.Sprintf("%s/%s", namespace, name),
			Namespace: namespace,
			Name:      name,
		},
		Description:        description,
		Versions:           versions,
		Warnings:           warnings,
		IsBlocked:          false,
		Popularity:         stats.Stars,
		ForkCount:          stats.Forks,
		UpstreamPopularity: 0, // Will be set below if this is a fork
		UpstreamForkCount:  0, // Will be set below if this is a fork
	}

	// Add fork information if applicable
	if repo.IsFork && repo.ParentOrganisation != nil && repo.ParentName != nil {
		// Create fork_of address - extract provider name from parent repo name
		parentProviderName := *repo.ParentName
		if len(parentProviderName) > 19 && parentProviderName[:19] == "terraform-provider-" {
			parentProviderName = parentProviderName[19:] // Remove "terraform-provider-" prefix
		}

		index.ForkOf = &ProviderAddr{
			Display:   fmt.Sprintf("%s/%s", *repo.ParentOrganisation, parentProviderName),
			Namespace: *repo.ParentOrganisation,
			Name:      parentProviderName,
		}

		// Set fork_of_link to GitHub URL
		githubURL := fmt.Sprintf("https://github.com/%s/%s", *repo.ParentOrganisation, *repo.ParentName)
		index.ForkOfLink = &githubURL

		// Query upstream repository stats
		upstreamStats, err := queryLatestRepositoryStats(ctx, db, *repo.ParentOrganisation, *repo.ParentName)
		if err == nil {
			index.UpstreamPopularity = upstreamStats.Stars
			index.UpstreamForkCount = upstreamStats.Forks
		}
		// If upstream stats query fails, we keep the zero values
	}

	return index, nil
}

// RebuildGlobalModuleIndex rebuilds the entire global module index from the database
func RebuildGlobalModuleIndex(ctx context.Context, db *pgxpool.Pool) (*GlobalModuleIndex, error) {
	ctx, span := telemetry.Tracer().Start(ctx, "index.rebuild_global_modules")
	defer span.End()

	// Single query that fetches all module data with JOINs
	query := `
		WITH latest_stats AS (
			SELECT DISTINCT ON (repo_organisation, repo_name)
				repo_organisation, repo_name, stars, forks
			FROM repository_stats
			ORDER BY repo_organisation, repo_name, recorded_at DESC
		),
		module_versions_agg AS (
			SELECT
				module_namespace,
				module_name,
				module_target,
				array_agg(version ORDER BY safe_to_semver(version) DESC) as versions,
				array_agg(discovered_at ORDER BY safe_to_semver(version) DESC) as discovered_dates
			FROM module_versions
			GROUP BY module_namespace, module_name, module_target
		)
		SELECT
			mv.module_namespace,
			mv.module_name,
			mv.module_target,
			mv.versions,
			mv.discovered_dates,

			COALESCE(s.stars, 0) as stars,
			COALESCE(s.forks, 0) as forks,

			r.description,
			COALESCE(r.is_fork, false) as is_fork,
			r.parent_organisation,
			r.parent_name,

			COALESCE(us.stars, 0) as upstream_stars,
			COALESCE(us.forks, 0) as upstream_forks
		FROM module_versions_agg mv
		-- repositories.name stores the full GitHub repo name (e.g. "terraform-aws-vpc"),
		-- while module_versions stores the short name (e.g. "vpc") and target (e.g. "aws") separately
		LEFT JOIN repositories r
			ON r.organisation = mv.module_namespace
			AND r.name = 'terraform-' || mv.module_target || '-' || mv.module_name
		LEFT JOIN latest_stats s
			ON s.repo_organisation = mv.module_namespace
			AND s.repo_name = 'terraform-' || mv.module_target || '-' || mv.module_name
		LEFT JOIN latest_stats us 
			ON us.repo_organisation = r.parent_organisation
			AND us.repo_name = r.parent_name
		ORDER BY mv.module_namespace, mv.module_name, mv.module_target`

	rows, err := db.Query(ctx, query)
	if err != nil {
		return nil, fmt.Errorf("failed to query modules: %w", err)
	}
	defer rows.Close()

	var modules []ModuleEntry
	for rows.Next() {
		var (
			namespace          string
			name               string
			target             string
			versions           []string
			discoveredDates    []time.Time
			stars              int
			forks              int
			description        *string
			isFork             bool
			parentOrganisation *string
			parentName         *string
			upstreamStars      int
			upstreamForks      int
		)

		if err := rows.Scan(
			&namespace,
			&name,
			&target,
			&versions,
			&discoveredDates,
			&stars,
			&forks,
			&description,
			&isFork,
			&parentOrganisation,
			&parentName,
			&upstreamStars,
			&upstreamForks,
		); err != nil {
			return nil, fmt.Errorf("failed to scan module row: %w", err)
		}

		if len(versions) == 0 {
			continue
		}

		// Get description
		desc := ""
		if description != nil {
			desc = *description
		}

		// Get published date for latest version
		publishedAt := time.Time{}
		if len(discoveredDates) > 0 {
			publishedAt = discoveredDates[0]
		}

		entry := ModuleEntry{
			Addr: ModuleAddr{
				Display:   fmt.Sprintf("%s/%s/%s", namespace, name, target),
				Namespace: namespace,
				Name:      name,
				Target:    target,
			},
			Description:   desc,
			LatestVersion: versions[0],
			PublishedAt:   publishedAt,
		}

		modules = append(modules, entry)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating module rows: %w", err)
	}

	return &GlobalModuleIndex{Modules: modules}, nil
}

// RebuildGlobalProviderIndex rebuilds the entire global provider index from the database
func RebuildGlobalProviderIndex(ctx context.Context, db *pgxpool.Pool) (*GlobalProviderIndex, error) {
	ctx, span := telemetry.Tracer().Start(ctx, "index.rebuild_global_providers")
	defer span.End()

	// Single query that fetches all provider data with JOINs
	query := `
		WITH latest_stats AS (
			SELECT DISTINCT ON (repo_organisation, repo_name)
				repo_organisation, repo_name, stars, forks
			FROM repository_stats
			ORDER BY repo_organisation, repo_name, recorded_at DESC
		),
		provider_versions_agg AS (
			SELECT
				provider_namespace,
				provider_name,
				array_agg(version ORDER BY safe_to_semver(version) DESC) as versions,
				array_agg(discovered_at ORDER BY safe_to_semver(version) DESC) as discovered_dates
			FROM provider_versions
			GROUP BY provider_namespace, provider_name
		)
		SELECT
			pv.provider_namespace,
			pv.provider_name,
			pv.versions,
			pv.discovered_dates,

			COALESCE(s.stars, 0) as stars,
			COALESCE(s.forks, 0) as forks,

			r.description,
			COALESCE(r.is_fork, false) as is_fork,
			r.parent_organisation,
			r.parent_name,
			pr.warnings,

			COALESCE(us.stars, 0) as upstream_stars,
			COALESCE(us.forks, 0) as upstream_forks
		FROM provider_versions_agg pv
		-- repositories.name stores the full GitHub repo name (e.g. "terraform-provider-aws"),
		-- while provider_versions.provider_name stores the short name (e.g. "aws")
		LEFT JOIN repositories r ON r.organisation = pv.provider_namespace
			AND r.name = 'terraform-provider-' || pv.provider_name
		LEFT JOIN latest_stats s ON s.repo_organisation = pv.provider_namespace
			AND s.repo_name = 'terraform-provider-' || pv.provider_name
		LEFT JOIN providers pr ON pr.namespace = pv.provider_namespace
			AND pr.name = pv.provider_name
		LEFT JOIN latest_stats us ON us.repo_organisation = r.parent_organisation
			AND us.repo_name = r.parent_name
		ORDER BY pv.provider_namespace, pv.provider_name`

	rows, err := db.Query(ctx, query)
	if err != nil {
		return nil, fmt.Errorf("failed to query providers: %w", err)
	}
	defer rows.Close()

	var providers []ProviderEntry
	for rows.Next() {
		var (
			namespace          string
			name               string
			versions           []string
			discoveredDates    []time.Time
			stars              int
			forks              int
			description        *string
			isFork             bool
			parentOrganisation *string
			parentName         *string
			warnings           []string
			upstreamStars      int
			upstreamForks      int
		)

		if err := rows.Scan(
			&namespace,
			&name,
			&versions,
			&discoveredDates,
			&stars,
			&forks,
			&description,
			&isFork,
			&parentOrganisation,
			&parentName,
			&warnings,
			&upstreamStars,
			&upstreamForks,
		); err != nil {
			return nil, fmt.Errorf("failed to scan provider row: %w", err)
		}

		if len(versions) == 0 {
			continue
		}

		// Build version info list
		versionInfos := make([]VersionInfo, len(versions))
		for i, v := range versions {
			var published *time.Time
			if i < len(discoveredDates) {
				published = &discoveredDates[i]
			}
			versionInfos[i] = VersionInfo{ID: v, Published: published}
		}

		// Get description
		desc := ""
		if description != nil {
			desc = *description
		}

		// Get published date for latest version
		publishedAt := time.Time{}
		if len(discoveredDates) > 0 {
			publishedAt = discoveredDates[0]
		}

		// Generate GitHub repository URL
		repoURL := fmt.Sprintf("https://github.com/%s/terraform-provider-%s", namespace, name)

		entry := ProviderEntry{
			Addr: ProviderAddr{
				Display:   fmt.Sprintf("%s/%s", namespace, name),
				Namespace: namespace,
				Name:      name,
			},
			Description:   desc,
			LatestVersion: versions[0],
			PublishedAt:   publishedAt,
			Link:          repoURL,
			Warnings:      warnings,
			Popularity:    stars,
			ForkCount:     forks,
		}

		// Add fork information if applicable
		if isFork && parentOrganisation != nil && parentName != nil {
			parentProviderName := *parentName
			if len(parentProviderName) > 19 && parentProviderName[:19] == "terraform-provider-" {
				parentProviderName = parentProviderName[19:]
			}

			entry.ForkOf = &ProviderAddr{
				Display:   fmt.Sprintf("%s/%s", *parentOrganisation, parentProviderName),
				Namespace: *parentOrganisation,
				Name:      parentProviderName,
			}

			githubURL := fmt.Sprintf("https://github.com/%s/%s", *parentOrganisation, *parentName)
			entry.ForkOfLink = &githubURL
		}

		providers = append(providers, entry)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating provider rows: %w", err)
	}

	return &GlobalProviderIndex{Providers: providers}, nil
}
