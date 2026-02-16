package repository

import (
	"context"
	"fmt"
	"log/slog"

	"github.com/jackc/pgx/v5/pgxpool"
)

// SyncRepositoryMetadata fetches and stores complete GitHub metadata for a repository.
// This includes stats (stars, forks, watchers), metadata (description, language, archived),
// and fork information (is_fork, parent repository).
func SyncRepositoryMetadata(ctx context.Context, pool *pgxpool.Pool, githubClient *Client, org, name string) error {
	if githubClient == nil {
		return fmt.Errorf("github client is required for repository sync")
	}

	slog.DebugContext(ctx, "Syncing repository metadata from GitHub",
		"repository", fmt.Sprintf("%s/%s", org, name))

	// Fetch repository metadata from GitHub API
	metadata, err := githubClient.GetRepositoryMetadata(ctx, org, name)
	if err != nil {
		return fmt.Errorf("failed to fetch repository metadata for %s/%s: %w", org, name, err)
	}

	// If this is a fork, ensure the parent repository exists first
	if metadata.IsFork && metadata.ParentOwner != "" && metadata.ParentName != "" {
		slog.DebugContext(ctx, "Repository is a fork, ensuring parent repository exists",
			"repository", fmt.Sprintf("%s/%s", org, name),
			"parent", fmt.Sprintf("%s/%s", metadata.ParentOwner, metadata.ParentName))

		// Insert parent repository if it doesn't exist (this will be a no-op if it already exists)
		parentQuery := `
			INSERT INTO repositories (organisation, name)
			VALUES ($1, $2)
			ON CONFLICT (organisation, name) DO NOTHING`

		_, err = pool.Exec(ctx, parentQuery, metadata.ParentOwner, metadata.ParentName)
		if err != nil {
			return fmt.Errorf("failed to ensure parent repository exists for %s/%s: %w", org, name, err)
		}
	}

	// Update/create repository metadata BEFORE storing stats (to satisfy foreign key constraint)
	err = UpdateRepositoryMetadata(ctx, pool, metadata)
	if err != nil {
		return fmt.Errorf("failed to update repository metadata for %s/%s: %w", org, name, err)
	}

	// Store repository statistics (stars, forks, watchers, subscribers, topics)
	// This must come AFTER UpdateRepositoryMetadata because of foreign key constraints
	err = StoreRepositoryStats(ctx, pool, metadata)
	if err != nil {
		return fmt.Errorf("failed to store repository stats for %s/%s: %w", org, name, err)
	}

	// Handle repository redirects if detected
	if metadata.IsRedirect && metadata.ActualOwner != "" && metadata.ActualName != "" {
		slog.InfoContext(ctx, "Repository redirect detected",
			"requested", fmt.Sprintf("%s/%s", metadata.Owner, metadata.Name),
			"actual", fmt.Sprintf("%s/%s", metadata.ActualOwner, metadata.ActualName))

		// Note: StoreRepositoryRedirect function would need to be implemented
		// For now, we log the redirect information
		slog.DebugContext(ctx, "Repository redirect info logged",
			"from", fmt.Sprintf("%s/%s", metadata.Owner, metadata.Name),
			"to", fmt.Sprintf("%s/%s", metadata.ActualOwner, metadata.ActualName))
	}

	slog.DebugContext(ctx, "Successfully synced repository metadata",
		"repository", fmt.Sprintf("%s/%s", org, name),
		"stars", metadata.Stars,
		"forks", metadata.Forks,
		"is_fork", metadata.IsFork)

	return nil
}

