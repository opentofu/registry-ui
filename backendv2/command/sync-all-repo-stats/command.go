// Package syncallrepostats implements the sync-all-repo-stats CLI command,
// which collects GitHub stats (stars, forks, watchers, open issues) via the
// GraphQL API and appends a time-series row per repository.
package syncallrepostats

import (
	"context"
	"fmt"
	"log/slog"
	"time"

	"github.com/google/uuid"
	"github.com/urfave/cli/v3"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/codes"

	"github.com/opentofu/registry-ui/pkg/config"
	"github.com/opentofu/registry-ui/pkg/repository"
	"github.com/opentofu/registry-ui/pkg/telemetry"
)

const (
	// defaultBatchSize is kept at 50 because the richer query (topics + open issuers)
	// exceeds GitHub's per-query resource budget at 100 repos,
	// causing repos to be silently dropped from a batch.
	defaultBatchSize    = 50
	defaultPointReserve = 500
	defaultStaleAfter   = 12 * time.Hour
	// defaultMaxAttempts retries each GraphQL request a few times because GitHub
	// is flakey and intermittently returns secondary rate limits.
	defaultMaxAttempts = 5
)

func NewCommand() *cli.Command {
	return &cli.Command{
		Name:  "sync-all-repo-stats",
		Usage: "Sync GitHub stats (stars, forks, watchers, open issues) for all repositories via the GraphQL API",
		Flags: []cli.Flag{
			&cli.IntFlag{
				Name:  "batch-size",
				Value: defaultBatchSize,
				Usage: "Number of repositories to fetch per GraphQL request",
			},
			&cli.IntFlag{
				Name:  "point-reserve",
				Value: defaultPointReserve,
				Usage: "Exit with a non-zero status when remaining GraphQL points drop to this value",
			},
			&cli.DurationFlag{
				Name:  "stale-after",
				Value: defaultStaleAfter,
				Usage: "Only sync repositories with no stats datapoint newer than this (e.g. 12h). Use 0 to sync all",
			},
			&cli.IntFlag{
				Name:  "max-attempts",
				Value: defaultMaxAttempts,
				Usage: "Number of times to attempt each GraphQL request before giving up (GitHub is flakey)",
			},
		},
		Action: func(ctx context.Context, cmd *cli.Command) error {
			return run(ctx, cmd)
		},
	}
}

func run(ctx context.Context, cmd *cli.Command) error {
	cfg := config.FromCLI(cmd)
	ctx, span := telemetry.Tracer().Start(ctx, "cmd.sync_all_repo_stats")
	defer span.End()

	batchJobID := uuid.New().String()
	batchSize := cmd.Int("batch-size")
	pointReserve := cmd.Int("point-reserve")
	staleAfter := cmd.Duration("stale-after")
	maxAttempts := cmd.Int("max-attempts")

	span.SetAttributes(
		attribute.String(telemetry.BatchJobIDKey, batchJobID),
		attribute.String(telemetry.BatchJobNameKey, "sync-all-repo-stats"),
		attribute.Int("batch_size", batchSize),
		attribute.Int("point_reserve", pointReserve),
		attribute.Float64("stale_after_seconds", staleAfter.Seconds()),
		attribute.Int("max_attempts", maxAttempts),
	)

	if batchSize <= 0 {
		batchSize = defaultBatchSize
	}
	if maxAttempts <= 0 {
		maxAttempts = defaultMaxAttempts
	}

	slog.InfoContext(ctx, "Starting repository stats sync",
		"batch_size", batchSize, "point_reserve", pointReserve,
		"stale_after", staleAfter.String(), "max_attempts", maxAttempts)

	pool, err := cfg.DB.GetPool(ctx)
	if err != nil {
		span.RecordError(err)
		span.SetStatus(codes.Error, err.Error())
		return fmt.Errorf("failed to initialize database pool: %w", err)
	}
	defer pool.Close()

	// Connect to the db and get a list of the repositories we're tracking
	githubClient := repository.NewClient(ctx, &cfg.GitHub)

	repos, err := repository.ListRepositoriesForStatsSync(ctx, pool, staleAfter)
	if err != nil {
		span.RecordError(err)
		span.SetStatus(codes.Error, err.Error())
		return fmt.Errorf("failed to list repositories: %w", err)
	}

	if len(repos) == 0 {
		slog.InfoContext(ctx, "No repositories need a stats sync", "stale_after", staleAfter.String())
		return nil
	}

	span.SetAttributes(attribute.Int("repositories.total", len(repos)))
	slog.InfoContext(ctx, "Repositories needing a stats sync", "count", len(repos), "stale_after", staleAfter.String())

	var syncedCount, failedCount int
	for start := 0; start < len(repos); start += batchSize {
		end := min(start+batchSize, len(repos))
		batch := repos[start:end]

		stats, rl, err := githubClient.GetRepositoryStatsBatch(ctx, batch, maxAttempts)
		if err != nil {
			span.RecordError(err)
			span.SetStatus(codes.Error, err.Error())
			return fmt.Errorf("failed to fetch stats for batch [%d:%d]: %w", start, end, err)
		}

		if err := repository.StoreRepositoryStatsBatch(ctx, pool, stats); err != nil {
			span.RecordError(err)
			span.SetStatus(codes.Error, err.Error())
			return fmt.Errorf("failed to store stats for batch [%d:%d]: %w", start, end, err)
		}

		syncedCount += len(stats)
		failedCount += len(batch) - len(stats)

		slog.InfoContext(ctx, "Synced repository stats batch",
			"from", start, "to", end, "of", len(repos),
			"fetched", len(stats), "points_remaining", rl.Remaining)

		// safety net: if the GraphQL budget approaches the reserve, fail fast
		// with a non-zero exit code rather than starving any co-running job.
		if rl.Remaining > 0 && rl.Remaining <= pointReserve {
			err := fmt.Errorf("GraphQL rate limit near floor: remaining %d <= reserve %d, reset_at %s", rl.Remaining, pointReserve, rl.ResetAt)
			span.RecordError(err)
			span.SetStatus(codes.Error, err.Error())
			slog.ErrorContext(ctx, "GraphQL rate limit near floor, aborting",
				"remaining", rl.Remaining, "reserve", pointReserve, "reset_at", rl.ResetAt)
			return err
		}
	}

	span.SetAttributes(
		attribute.Int("repositories.synced", syncedCount),
		attribute.Int("repositories.failed", failedCount),
	)
	slog.InfoContext(ctx, "Repository stats sync completed",
		"total", len(repos), "synced", syncedCount, "failed", failedCount)

	return nil
}
