// Package syncrepostats implements a CLI command to sync repository statistics.
// This is intended to be used for testing or backfilling stats for existing repositories.
package syncrepostats

import (
	"context"
	"fmt"
	"log/slog"

	"github.com/urfave/cli/v3"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/codes"

	"github.com/opentofu/registry-ui/pkg/config"
	"github.com/opentofu/registry-ui/pkg/repository"
	"github.com/opentofu/registry-ui/pkg/telemetry"
)

func NewCommand() *cli.Command {
	return &cli.Command{
		Name:  "sync-repo-stats",
		Usage: "Sync repository statistics (stars, forks, etc.) from GitHub",
		Flags: []cli.Flag{
			&cli.StringFlag{
				Name:     "namespace",
				Aliases:  []string{"n"},
				Usage:    "GitHub organisation (e.g., hashicorp)",
				Required: true,
			},
			&cli.StringFlag{
				Name:     "name",
				Usage:    "GitHub repository name (e.g., terraform-provider-aws)",
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
	ctx, span := telemetry.Tracer().Start(ctx, "cmd.sync_repo_stats")
	defer span.End()

	// Get command flags
	org := cmd.String("namespace")
	repoName := cmd.String("name")

	span.SetAttributes(
		attribute.String("repo.organisation", org),
		attribute.String("repo.name", repoName),
	)

	slog.InfoContext(ctx, "Starting repository stats sync",
		"organisation", org, "repo", repoName)

	// Connect to database
	pool, err := cfg.DB.GetPool(ctx)
	if err != nil {
		span.RecordError(err)
		span.SetStatus(codes.Error, err.Error())
		slog.ErrorContext(ctx, "Failed to connect to database", "error", err)
		return fmt.Errorf("failed to connect to database: %w", err)
	}
	defer pool.Close()

	// Create GitHub client
	githubClient := repository.NewClient(ctx, &cfg.GitHub)

	// Sync repository metadata directly using org/repo name
	err = repository.SyncRepositoryMetadata(ctx, pool, githubClient, org, repoName)
	if err != nil {
		span.RecordError(err)
		span.SetStatus(codes.Error, err.Error())
		return fmt.Errorf("failed to sync repository stats for %s/%s: %w", org, repoName, err)
	}

	slog.InfoContext(ctx, "Successfully synced repository stats",
		"organisation", org, "repo", repoName)

	return nil
}
