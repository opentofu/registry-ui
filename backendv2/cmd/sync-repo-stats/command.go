// Package syncrepostats implements a CLI command to sync repository statistics for providers.
// This is intended to be used for testing or backfilling stats for existing providers.
package syncrepostats

import (
	"context"
	"fmt"
	"log/slog"

	"github.com/urfave/cli/v3"
	"go.opentelemetry.io/otel/attribute"

	"github.com/opentofu/registry-ui/pkg/config"
	"github.com/opentofu/registry-ui/pkg/provider/storage"
	"github.com/opentofu/registry-ui/pkg/repository"
	"github.com/opentofu/registry-ui/pkg/telemetry"
)

func NewCommand() *cli.Command {
	return &cli.Command{
		Name:  "sync-repo-stats",
		Usage: "Sync repository statistics for a specific provider (stars, forks, etc.)",
		Flags: []cli.Flag{
			&cli.StringFlag{
				Name:     "namespace",
				Aliases:  []string{"n"},
				Usage:    "Provider namespace (e.g., hashicorp)",
				Required: true,
			},
			&cli.StringFlag{
				Name:     "name",
				Usage:    "Provider name (e.g., aws)",
				Required: true,
			},
		},
		Action: func(ctx context.Context, cmd *cli.Command) error {
			return run(ctx, cmd)
		},
	}
}

func run(ctx context.Context, cmd *cli.Command) error {
	cfg := config.FromCLI(cmd)
	ctx, span := telemetry.Tracer().Start(ctx, "sync-repo-stats")
	defer span.End()

	// Get command flags
	namespace := cmd.String("namespace")
	name := cmd.String("name")

	slog.InfoContext(ctx, "Starting repository stats sync for provider",
		"namespace", namespace, "name", name)

	// Connect to database
	pool, err := cfg.DB.GetPool(ctx)
	if err != nil {
		span.RecordError(err)
		return fmt.Errorf("failed to connect to database: %w", err)
	}
	defer pool.Close()

	// Create GitHub client
	githubClient := repository.NewClient(&cfg.GitHub)

	// Get provider repository info from database
	provider, err := storage.GetProvider(ctx, pool, namespace, name)
	if err != nil {
		span.RecordError(err)
		return fmt.Errorf("failed to get provider: %w", err)
	}

	span.SetAttributes(
		attribute.String("provider.namespace", provider.Namespace),
		attribute.String("provider.name", provider.Name),
		attribute.String("repo.organisation", provider.RepoOrganisation),
		attribute.String("repo.name", provider.RepoName),
	)

	slog.InfoContext(ctx, "Syncing repository stats",
		"provider", fmt.Sprintf("%s/%s", provider.Namespace, provider.Name),
		"repository", fmt.Sprintf("%s/%s", provider.RepoOrganisation, provider.RepoName))

	// Fetch repository metadata from GitHub
	metadata, err := githubClient.GetRepositoryMetadata(ctx, provider.RepoOrganisation, provider.RepoName)
	if err != nil {
		span.RecordError(err)
		return fmt.Errorf("failed to fetch repository metadata for %s/%s: %w",
			provider.RepoOrganisation, provider.RepoName, err)
	}

	// Store repository stats in database
	err = repository.StoreRepositoryStats(ctx, pool, metadata)
	if err != nil {
		span.RecordError(err)
		return fmt.Errorf("failed to store repository stats: %w", err)
	}

	// Update repositories table with latest metadata
	err = repository.UpdateRepositoryMetadata(ctx, pool, metadata)
	if err != nil {
		span.RecordError(err)
		return fmt.Errorf("failed to update repository metadata: %w", err)
	}

	slog.InfoContext(ctx, "Successfully synced repository stats",
		"provider", fmt.Sprintf("%s/%s", provider.Namespace, provider.Name),
		"repository", fmt.Sprintf("%s/%s", provider.RepoOrganisation, provider.RepoName),
		"stars", metadata.Stars,
		"forks", metadata.Forks)

	return nil
}
