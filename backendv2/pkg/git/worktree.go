package git

import (
	"context"
	"fmt"
	"path/filepath"

	"go.opentelemetry.io/otel/attribute"

	"github.com/opentofu/registry-ui/pkg/telemetry"
)

// WorktreeManager handles worktree lifecycle for a specific tag/version
type WorktreeManager struct {
	repo         *Repo
	tag          string
	worktreePath string
	repoURL      string
}

// NewWorktreeManager creates a worktree for the specified tag and returns a manager
func NewWorktreeManager(ctx context.Context, repoURL, localPath, tag string) (*WorktreeManager, error) {
	ctx, span := telemetry.Tracer().Start(ctx, "git.new_worktree_manager")
	defer span.End()

	span.SetAttributes(
		attribute.String("git.repo_url", repoURL),
		attribute.String("git.local_path", localPath),
		attribute.String("git.tag", tag),
	)

	repo, err := GetRepo(repoURL, localPath)
	if err != nil {
		span.RecordError(err)
		return nil, fmt.Errorf("failed to get repository: %w", err)
	}

	if err := repo.EnsureCloned(ctx); err != nil {
		span.RecordError(err)
		return nil, fmt.Errorf("failed to clone repository: %w", err)
	}

	// Fetch all tags from remote
	if err := repo.FetchTags(ctx); err != nil {
		span.RecordError(err)
		return nil, fmt.Errorf("failed to fetch tags: %w", err)
	}

	// Create worktree for the specific tag (use as-is from registry)
	worktreePath := filepath.Join(localPath, ".tofu-worktrees", tag)
	if err := repo.AddWorktree(ctx, tag, worktreePath); err != nil {
		span.RecordError(err)
		return nil, fmt.Errorf("failed to create worktree for tag %s: %w", tag, err)
	}

	return &WorktreeManager{
		repo:         repo,
		tag:          tag,
		worktreePath: worktreePath,
		repoURL:      repoURL,
	}, nil
}

// Path returns the filesystem path to the worktree
func (wm *WorktreeManager) Path() string {
	return wm.worktreePath
}

// Cleanup removes the worktree
func (wm *WorktreeManager) Cleanup(ctx context.Context) error {
	ctx, span := telemetry.Tracer().Start(ctx, "git.cleanup_worktree")
	defer span.End()

	span.SetAttributes(
		attribute.String("git.repo_url", wm.repoURL),
		attribute.String("git.tag", wm.tag),
		attribute.String("git.worktree_path", wm.worktreePath),
	)

	if err := wm.repo.RemoveWorktree(ctx, wm.tag); err != nil {
		span.RecordError(err)
		return fmt.Errorf("failed to remove worktree for tag %s: %w", wm.tag, err)
	}

	return nil
}

// CheckoutVersionForScraping creates a worktree for a tag and returns the directory path and cleanup function
func CheckoutVersionForScraping(ctx context.Context, repoURL, localPath, tag string) (string, func(), error) {
	wm, err := NewWorktreeManager(ctx, repoURL, localPath, tag)
	if err != nil {
		return "", nil, err
	}

	cleanup := func() {
		if cleanupErr := wm.Cleanup(ctx); cleanupErr != nil {
			// Log the error but don't fail
			_, span := telemetry.Tracer().Start(ctx, "git.cleanup")
			span.RecordError(cleanupErr)
			span.End()
		}
	}

	return wm.Path(), cleanup, nil
}
