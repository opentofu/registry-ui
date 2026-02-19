package git

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"sync"
	"time"

	"github.com/go-git/go-git/v5"
	"github.com/go-git/go-git/v5/config"
	"github.com/go-git/go-git/v5/plumbing"
	"github.com/go-git/go-git/v5/plumbing/transport"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/trace"

	"github.com/opentofu/registry-ui/pkg/telemetry"
)

// Repo represents a git repository with worktree management capabilities
type Repo struct {
	URL       string
	LocalPath string

	// Internal state
	cloned     bool
	repository *git.Repository
	worktrees  sync.Map // ref -> path mapping for worktrees (thread-safe)
}

// requireCloned requires that the repository is cloned and ready for operations.
// Records error on span and logs if not cloned.
func (r *Repo) requireCloned(ctx context.Context, span trace.Span, action string) error {
	if !r.cloned || r.repository == nil {
		err := fmt.Errorf("repository not cloned")
		span.RecordError(err)
		slog.ErrorContext(ctx, fmt.Sprintf("Cannot %s, repository not cloned", action), "url", r.URL, "error", err)
		return err
	}
	return nil
}

// GetRepo creates or returns an existing Repo instance
func GetRepo(url, localPath string) (*Repo, error) {
	if url == "" {
		return nil, fmt.Errorf("repository URL cannot be empty")
	}
	if localPath == "" {
		return nil, fmt.Errorf("local path cannot be empty")
	}

	localPath = filepath.Clean(localPath)
	if info, err := os.Stat(localPath); err == nil {
		if !info.IsDir() {
			return nil, fmt.Errorf("local path %s exists but is not a directory", localPath)
		}
	}

	repo := &Repo{
		URL:       url,
		LocalPath: localPath,
	}

	if _, err := os.Stat(filepath.Join(localPath, ".git")); err == nil {
		repository, err := git.PlainOpen(localPath)
		if err != nil {
			return nil, fmt.Errorf("failed to open existing repository at %s: %w", localPath, err)
		}
		repo.repository = repository
		repo.cloned = true
	}

	return repo, nil
}

// EnsureCloned clones the repository if it hasn't been cloned already
func (r *Repo) EnsureCloned(ctx context.Context) error {
	if r.cloned && r.repository != nil {
		slog.DebugContext(ctx, "Repository already cloned", "url", r.URL, "path", r.LocalPath)
		return nil
	}

	ctx, span := telemetry.Tracer().Start(ctx, "git.clone")
	defer span.End()

	span.SetAttributes(
		attribute.String("git.url", r.URL),
		attribute.String("git.local_path", r.LocalPath),
	)

	slog.DebugContext(ctx, "Cloning repository", "url", r.URL, "path", r.LocalPath)

	if err := os.MkdirAll(r.LocalPath, 0o755); err != nil {
		span.RecordError(err)
		return fmt.Errorf("failed to create directory %s: %w", r.LocalPath, err)
	}

	cloneOptions := &git.CloneOptions{
		URL:        r.URL,
		Progress:   nil,
		NoCheckout: true, // Don't checkout working tree
	}

	repository, err := git.PlainCloneContext(ctx, r.LocalPath, false, cloneOptions)
	if err != nil {
		span.RecordError(err)

		if errors.Is(err, transport.ErrEmptyRemoteRepository) {
			if removeErr := os.RemoveAll(r.LocalPath); removeErr != nil {
				return fmt.Errorf("failed to clean up after empty repository %s: %w (original: %w)", r.URL, removeErr, err)
			}
			return fmt.Errorf("remote repository %s is empty", r.URL)
		}
		if errors.Is(err, transport.ErrRepositoryNotFound) {
			if removeErr := os.RemoveAll(r.LocalPath); removeErr != nil {
				return fmt.Errorf("failed to clean up after missing repository %s: %w (original: %w)", r.URL, removeErr, err)
			}
			return fmt.Errorf("repository %s not found", r.URL)
		}
		if errors.Is(err, git.ErrRepositoryAlreadyExists) {
			slog.DebugContext(ctx, "Repository already exists, opening existing", "path", r.LocalPath)
			repository, err = git.PlainOpen(r.LocalPath)
			if err != nil {
				return fmt.Errorf("repository exists but cannot be opened: %w", err)
			}
		} else {
			// Unknown error - clean up and return
			if removeErr := os.RemoveAll(r.LocalPath); removeErr != nil {
				return fmt.Errorf("failed to clean up after clone error for %s: %w (original: %w)", r.URL, removeErr, err)
			}
			return fmt.Errorf("failed to clone repository %s, unknown error: %w", r.URL, err)
		}
	}

	r.repository = repository
	r.cloned = true

	// Clean up any stale worktrees from previous runs
	pruneCmd := exec.CommandContext(ctx, "git", "worktree", "prune")
	pruneCmd.Dir = r.LocalPath
	if err := pruneCmd.Run(); err != nil {
		slog.DebugContext(ctx, "Failed to prune worktrees", "url", r.URL, "error", err)
	}

	slog.DebugContext(ctx, "Successfully cloned repository", "url", r.URL, "path", r.LocalPath)
	return nil
}

// FetchTags fetches all tags from the remote repository
func (r *Repo) FetchTags(ctx context.Context) error {
	ctx, span := telemetry.Tracer().Start(ctx, "git.fetch_tags")
	defer span.End()

	span.SetAttributes(
		attribute.String("git.url", r.URL),
		attribute.String("git.local_path", r.LocalPath),
	)

	if err := r.requireCloned(ctx, span, "fetch tags"); err != nil {
		return err
	}

	slog.DebugContext(ctx, "Fetching all tags from remote", "url", r.URL)

	err := r.repository.FetchContext(ctx, &git.FetchOptions{
		RemoteName: "origin",
		RefSpecs: []config.RefSpec{
			"+refs/tags/*:refs/tags/*",
		},
	})

	// git.NoErrAlreadyUpToDate is not an error, it just means we already have all tags
	if err != nil && !errors.Is(err, git.NoErrAlreadyUpToDate) {
		span.RecordError(err)
		slog.ErrorContext(ctx, "Failed to fetch tags", "url", r.URL, "error", err)
		return fmt.Errorf("failed to fetch tags: %w", err)
	}

	slog.DebugContext(ctx, "Successfully fetched all tags", "url", r.URL)
	return nil
}

func (r *Repo) AddWorktree(ctx context.Context, ref, path string) error {
	ctx, span := telemetry.Tracer().Start(ctx, "git.add_worktree")
	defer span.End()

	span.SetAttributes(
		attribute.String("git.url", r.URL),
		attribute.String("git.ref", ref),
		attribute.String("git.worktree_path", path),
	)

	slog.DebugContext(ctx, "Creating git worktree", "url", r.URL, "ref", ref, "path", path)

	if err := r.requireCloned(ctx, span, "create worktree"); err != nil {
		return err
	}

	// Prune stale worktree registrations before attempting to add
	pruneCmd := exec.CommandContext(ctx, "git", "worktree", "prune")
	pruneCmd.Dir = r.LocalPath
	if err := pruneCmd.Run(); err != nil {
		slog.DebugContext(ctx, "Failed to prune worktrees", "url", r.URL, "error", err)
	}

	// Verify the ref exists in the repository
	if _, err := r.repository.ResolveRevision(plumbing.Revision(ref)); err != nil {
		span.RecordError(err)
		slog.ErrorContext(ctx, "Ref does not exist in repository", "url", r.URL, "ref", ref, "error", err)
		return fmt.Errorf("ref %s does not exist: %w", ref, err)
	}

	if existingPathRaw, exists := r.worktrees.Load(ref); exists {
		existingPath, ok := existingPathRaw.(string)
		if !ok {
			err := fmt.Errorf("invalid worktree path type in map for ref %s", ref)
			span.RecordError(err)
			return err
		}
		if _, err := os.Stat(existingPath); err == nil {
			slog.DebugContext(ctx, "Worktree already exists", "url", r.URL, "ref", ref, "path", existingPath)
			return nil
		}
		slog.DebugContext(ctx, "Removing stale worktree entry", "url", r.URL, "ref", ref, "old_path", existingPath)
		r.worktrees.Delete(ref)
	}

	path, err := filepath.Abs(path)
	if err != nil {
		span.RecordError(err)
		return fmt.Errorf("failed to get absolute path for %s: %w", path, err)
	}

	// Remove existing directory if it exists (handles stale worktrees from interrupted runs)
	if _, err := os.Stat(path); err == nil {
		slog.DebugContext(ctx, "Removing stale worktree directory", "url", r.URL, "ref", ref, "path", path)
		if err := os.RemoveAll(path); err != nil {
			rmErr := os.RemoveAll(path)
			if rmErr != nil {
				span.RecordError(rmErr)
				return fmt.Errorf("failed to remove stale worktree directory with fallback: %w", rmErr)
			}
			span.RecordError(err)
			return fmt.Errorf("failed to remove existing worktree directory %s: %w", path, err)
		}
	}

	// Create the worktree (--detach creates a detached HEAD)
	cmd := exec.CommandContext(ctx, "git", "worktree", "add", "--detach", path, ref)
	cmd.Dir = r.LocalPath
	if output, err := cmd.CombinedOutput(); err != nil {
		span.RecordError(err)
		// Don't add command output to span - can be very large and cause OTLP export failures
		slog.ErrorContext(ctx, "Failed to create worktree", "url", r.URL, "ref", ref, "path", path, "output", string(output), "error", err)
		return fmt.Errorf("failed to create worktree: %w", err)
	}

	r.worktrees.Store(ref, path)
	slog.DebugContext(ctx, "Successfully created worktree", "url", r.URL, "ref", ref, "path", path)
	return nil
}

func (r *Repo) RemoveWorktree(ctx context.Context, ref string) error {
	ctx, span := telemetry.Tracer().Start(ctx, "git.remove_worktree")
	defer span.End()

	span.SetAttributes(
		attribute.String("git.url", r.URL),
		attribute.String("git.ref", ref),
	)

	slog.DebugContext(ctx, "Removing git worktree", "url", r.URL, "ref", ref)

	p, exists := r.worktrees.Load(ref)
	if !exists {
		slog.DebugContext(ctx, "Worktree does not exist", "url", r.URL, "ref", ref)
		return nil
	}

	path := p.(string)

	span.SetAttributes(attribute.String("git.worktree_path", path))

	cmd := exec.CommandContext(ctx, "git", "worktree", "remove", path, "--force")
	cmd.Dir = r.LocalPath
	if output, err := cmd.CombinedOutput(); err != nil {
		span.RecordError(err)
		// Don't add command output to span - can be very large and cause OTLP export failures
		slog.WarnContext(ctx, "Failed to remove worktree (continuing anyway)", "url", r.URL, "ref", ref, "path", path, "output", string(output), "error", err)
	} else {
		slog.DebugContext(ctx, "Successfully removed worktree", "url", r.URL, "ref", ref, "path", path)
	}

	r.worktrees.Delete(ref)
	return nil
}

// GetTagDate returns the commit date for a git tag (works for both lightweight and annotated tags)
func (r *Repo) GetTagDate(ctx context.Context, tag string) (*time.Time, error) {
	ctx, span := telemetry.Tracer().Start(ctx, "git.get_tag_date")
	defer span.End()

	span.SetAttributes(
		attribute.String("git.url", r.URL),
		attribute.String("git.tag", tag),
	)

	if err := r.requireCloned(ctx, span, "get tag date"); err != nil {
		return nil, err
	}

	slog.DebugContext(ctx, "Getting tag commit date", "url", r.URL, "tag", tag)

	// Try to resolve the tag to a commit hash
	hash, err := r.repository.ResolveRevision(plumbing.Revision(tag))
	if err != nil {
		// Try with 'v' prefix if not found
		if !strings.HasPrefix(tag, "v") {
			slog.DebugContext(ctx, "Tag not found, trying with 'v' prefix", "tag", tag)
			hash, err = r.repository.ResolveRevision(plumbing.Revision("v" + tag))
			if err != nil {
				span.RecordError(err)
				slog.WarnContext(ctx, "Failed to resolve tag", "url", r.URL, "tag", tag, "error", err)
				return nil, fmt.Errorf("failed to resolve tag %s: %w", tag, err)
			}
		} else {
			span.RecordError(err)
			slog.WarnContext(ctx, "Failed to resolve tag", "url", r.URL, "tag", tag, "error", err)
			return nil, fmt.Errorf("failed to resolve tag %s: %w", tag, err)
		}
	}

	// Get the commit object
	commit, err := r.repository.CommitObject(*hash)
	if err != nil {
		span.RecordError(err)
		slog.ErrorContext(ctx, "Failed to get commit for tag", "url", r.URL, "tag", tag, "hash", hash.String(), "error", err)
		return nil, fmt.Errorf("failed to get commit for tag %s: %w", tag, err)
	}

	tagDate := commit.Committer.When
	slog.DebugContext(ctx, "Successfully retrieved tag commit date", "url", r.URL, "tag", tag, "date", tagDate)
	return &tagDate, nil
}
