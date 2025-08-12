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
	"github.com/go-git/go-git/v5/plumbing/transport"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
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

// Clone clones the repository if it hasn't been cloned already
func (r *Repo) Clone(ctx context.Context) error {
	if r.cloned && r.repository != nil {
		slog.DebugContext(ctx, "Repository already cloned", "url", r.URL, "path", r.LocalPath)
		return nil
	}

	tracer := otel.Tracer("opentofu-registry-backend")
	ctx, span := tracer.Start(ctx, "git.clone")
	defer span.End()

	span.SetAttributes(
		attribute.String("git.url", r.URL),
		attribute.String("git.local_path", r.LocalPath),
	)

	slog.InfoContext(ctx, "Cloning repository", "url", r.URL, "path", r.LocalPath)

	if err := os.MkdirAll(r.LocalPath, 0755); err != nil {
		span.RecordError(err)
		return fmt.Errorf("failed to create directory %s: %w", r.LocalPath, err)
	}

	cloneOptions := &git.CloneOptions{
		URL:          r.URL,
		Progress:     nil,
		Depth:        1,           // Shallow clone - only latest commit
		SingleBranch: false,       // Get all branches
		NoCheckout:   true,        // Don't checkout working tree
		Tags:         git.AllTags, // Fetch all tags
	}

	repository, err := git.PlainCloneContext(ctx, r.LocalPath, false, cloneOptions)
	if err != nil {
		span.RecordError(err)
		os.RemoveAll(r.LocalPath)

		if errors.Is(err, transport.ErrEmptyRemoteRepository) {
			return fmt.Errorf("remote repository %s is empty", r.URL)
		}
		if errors.Is(err, transport.ErrRepositoryNotFound) {
			return fmt.Errorf("repository %s not found", r.URL)
		}
		if errors.Is(err, git.ErrRepositoryAlreadyExists) {
			slog.DebugContext(ctx, "Repository already exists, opening existing", "path", r.LocalPath)
			repository, err = git.PlainOpen(r.LocalPath)
			if err != nil {
				return fmt.Errorf("repository exists but cannot be opened: %w", err)
			}
		} else {
			return fmt.Errorf("failed to clone repository %s: %w", r.URL, err)
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

	slog.InfoContext(ctx, "Successfully cloned repository", "url", r.URL, "path", r.LocalPath)
	return nil
}

func (r *Repo) AddWorktree(ctx context.Context, ref, path string) error {
	tracer := otel.Tracer("opentofu-registry-backend")
	ctx, span := tracer.Start(ctx, "git.add_worktree")
	defer span.End()

	span.SetAttributes(
		attribute.String("git.url", r.URL),
		attribute.String("git.ref", ref),
		attribute.String("git.worktree_path", path),
	)

	slog.InfoContext(ctx, "Creating git worktree", "url", r.URL, "ref", ref, "path", path)

	if !r.cloned {
		err := fmt.Errorf("repository not cloned")
		span.RecordError(err)
		slog.ErrorContext(ctx, "Cannot create worktree, repository not cloned", "url", r.URL, "ref", ref, "error", err)
		return err
	}

	if existingPathRaw, exists := r.worktrees.Load(ref); exists {
		existingPath := existingPathRaw.(string)
		if _, err := os.Stat(existingPath); err == nil {
			slog.DebugContext(ctx, "Worktree already exists", "url", r.URL, "ref", ref, "path", existingPath)
			return nil
		}
		slog.DebugContext(ctx, "Removing stale worktree entry", "url", r.URL, "ref", ref, "old_path", existingPath)
		r.worktrees.Delete(ref)
	}

	path = filepath.Clean(path)
	if !filepath.IsAbs(path) {
		abs, err := filepath.Abs(path)
		if err != nil {
			span.RecordError(err)
			return fmt.Errorf("failed to get absolute path for %s: %w", path, err)
		}
		path = abs
	}

	// Remove existing directory if it exists
	if _, err := os.Stat(path); err == nil {
		slog.DebugContext(ctx, "Removing existing worktree directory", "url", r.URL, "ref", ref, "path", path)
		if err := os.RemoveAll(path); err != nil {
			span.RecordError(err)
			return fmt.Errorf("failed to remove existing worktree directory %s: %w", path, err)
		}
	}

	cmd := exec.CommandContext(ctx, "git", "worktree", "add", "--force", path, ref)
	cmd.Dir = r.LocalPath
	if output, err := cmd.CombinedOutput(); err != nil {
		span.RecordError(err)
		span.SetAttributes(attribute.String("git.command_output", string(output)))
		slog.ErrorContext(ctx, "Failed to create worktree", "url", r.URL, "ref", ref, "path", path, "output", string(output), "error", err)
		return fmt.Errorf("failed to create worktree: %w", err)
	}

	r.worktrees.Store(ref, path)
	slog.InfoContext(ctx, "Successfully created worktree", "url", r.URL, "ref", ref, "path", path)
	return nil
}

// Update fetches all updates from the remote repository (branches and tags)
func (r *Repo) Update(ctx context.Context) error {
	tracer := otel.Tracer("opentofu-registry-backend")
	ctx, span := tracer.Start(ctx, "git.update")
	defer span.End()

	span.SetAttributes(
		attribute.String("git.url", r.URL),
		attribute.String("git.local_path", r.LocalPath),
	)

	slog.InfoContext(ctx, "Updating repository from remote", "url", r.URL, "path", r.LocalPath)

	if !r.cloned || r.repository == nil {
		err := fmt.Errorf("repository not cloned")
		span.RecordError(err)
		slog.ErrorContext(ctx, "Cannot update, repository not cloned", "url", r.URL, "error", err)
		return err
	}

	fetchOptions := &git.FetchOptions{
		Tags:  git.AllTags,
		Depth: 1, // Only fetch the latest commit for each ref
	}

	err := r.repository.FetchContext(ctx, fetchOptions)
	if err != nil && err != git.NoErrAlreadyUpToDate {
		span.RecordError(err)
		slog.ErrorContext(ctx, "Failed to update repository", "url", r.URL, "error", err)
		return fmt.Errorf("failed to update repository: %w", err)
	}

	if errors.Is(err, git.NoErrAlreadyUpToDate) {
		slog.DebugContext(ctx, "Repository already up to date", "url", r.URL)
	} else {
		slog.InfoContext(ctx, "Successfully updated repository", "url", r.URL)
	}

	span.SetAttributes(attribute.Bool("git.already_up_to_date", errors.Is(err, git.NoErrAlreadyUpToDate)))
	return nil
}

func (r *Repo) RemoveWorktree(ctx context.Context, ref string) error {
	tracer := otel.Tracer("opentofu-registry-backend")
	ctx, span := tracer.Start(ctx, "git.remove_worktree")
	defer span.End()

	span.SetAttributes(
		attribute.String("git.url", r.URL),
		attribute.String("git.ref", ref),
	)

	slog.InfoContext(ctx, "Removing git worktree", "url", r.URL, "ref", ref)

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
		span.SetAttributes(attribute.String("git.command_output", string(output)))
		slog.WarnContext(ctx, "Failed to remove worktree (continuing anyway)", "url", r.URL, "ref", ref, "path", path, "output", string(output), "error", err)
	} else {
		slog.InfoContext(ctx, "Successfully removed worktree", "url", r.URL, "ref", ref, "path", path)
	}

	r.worktrees.Delete(ref)
	return nil
}

// Cleanup removes all worktrees and cleans up the repository
func (r *Repo) Cleanup(ctx context.Context) error {
	tracer := otel.Tracer("opentofu-registry-backend")
	ctx, span := tracer.Start(ctx, "git.cleanup")
	defer span.End()

	span.SetAttributes(
		attribute.String("git.url", r.URL),
		attribute.String("git.local_path", r.LocalPath),
	)

	slog.InfoContext(ctx, "Cleaning up repository and all worktrees", "url", r.URL, "path", r.LocalPath)

	// I am not the biggest fan of sync.Map sometimes, but its way better than mutex locking everywhere
	r.worktrees.Range(func(re, p interface{}) bool {
		ref := re.(string)
		path := p.(string)
		span.SetAttributes(attribute.String("git.cleaning_ref", ref))
		slog.DebugContext(ctx, "Removing worktree during cleanup", "url", r.URL, "ref", ref, "path", path)
		cmd := exec.CommandContext(ctx, "git", "worktree", "remove", path, "--force")
		cmd.Dir = r.LocalPath
		if output, err := cmd.CombinedOutput(); err != nil {
			slog.WarnContext(ctx, "Failed to remove worktree during cleanup", "url", r.URL, "ref", ref, "path", path, "output", string(output), "error", err)
		}
		return true
	})

	r.worktrees = sync.Map{}

	if r.LocalPath != "" {
		slog.InfoContext(ctx, "Removing repository directory", "url", r.URL, "path", r.LocalPath)
		if err := os.RemoveAll(r.LocalPath); err != nil {
			span.RecordError(err)
			slog.ErrorContext(ctx, "Failed to remove repository directory", "url", r.URL, "path", r.LocalPath, "error", err)
			return fmt.Errorf("failed to remove repository directory: %w", err)
		}
	}

	r.repository = nil
	r.cloned = false

	slog.InfoContext(ctx, "Successfully cleaned up repository", "url", r.URL, "path", r.LocalPath)
	return nil
}

// GetTagDate returns the creation date of a git tag
func (r *Repo) GetTagDate(ctx context.Context, tag string) (*time.Time, error) {
	tracer := otel.Tracer("opentofu-registry-backend")
	ctx, span := tracer.Start(ctx, "git.get_tag_date")
	defer span.End()

	span.SetAttributes(
		attribute.String("git.url", r.URL),
		attribute.String("git.tag", tag),
	)

	// Check if repository directory exists
	if _, err := os.Stat(filepath.Join(r.LocalPath, ".git")); os.IsNotExist(err) {
		err := fmt.Errorf("repository directory does not exist at %s", r.LocalPath)
		span.RecordError(err)
		return nil, err
	}

	slog.DebugContext(ctx, "Getting tag creation date", "url", r.URL, "tag", tag)

	// Use git log to get tag date
	cmd := exec.CommandContext(ctx, "git", "log", "--format=%ai", "-1", tag)
	cmd.Dir = r.LocalPath

	output, err := cmd.Output()
	if err != nil {
		// Try with 'v' prefix if not found
		if !strings.HasPrefix(tag, "v") {
			slog.DebugContext(ctx, "Tag not found, trying with 'v' prefix", "tag", tag)
			return r.GetTagDate(ctx, "v"+tag)
		}
		span.RecordError(err)
		slog.WarnContext(ctx, "Failed to get tag date", "url", r.URL, "tag", tag, "error", err)
		return nil, fmt.Errorf("failed to get tag date for %s: %w", tag, err)
	}

	dateStr := strings.TrimSpace(string(output))
	if dateStr == "" {
		err := fmt.Errorf("empty date output for tag %s", tag)
		span.RecordError(err)
		return nil, err
	}

	// Parse the git date format (RFC2822 style)
	tagDate, err := time.Parse("2006-01-02 15:04:05 -0700", dateStr)
	if err != nil {
		span.RecordError(err)
		slog.ErrorContext(ctx, "Failed to parse tag date", "url", r.URL, "tag", tag, "date_str", dateStr, "error", err)
		return nil, fmt.Errorf("failed to parse tag date: %w", err)
	}

	slog.DebugContext(ctx, "Successfully retrieved tag date", "url", r.URL, "tag", tag, "date", tagDate)
	return &tagDate, nil
}
