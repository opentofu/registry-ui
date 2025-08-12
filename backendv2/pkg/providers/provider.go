package providers

import (
	"context"
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"strings"
	"time"

	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"

	"github.com/opentofu/registry-ui/pkg/config"
	"github.com/opentofu/registry-ui/pkg/git"
	"github.com/opentofu/registry-ui/pkg/license"
)

type Provider struct {
	config    *config.BackendConfig
	namespace string
	name      string
}

func NewProvider(cfg *config.BackendConfig, namespace, name string) *Provider {
	return &Provider{
		config:    cfg,
		namespace: namespace,
		name:      name,
	}
}

// GetTagCreationDate returns the creation date of a git tag for this provider
func (p *Provider) GetTagCreationDate(ctx context.Context, version string) (*time.Time, error) {
	repoURL := fmt.Sprintf("https://github.com/%s/terraform-provider-%s", p.namespace, p.name)
	localPath := filepath.Join(p.config.WorkDir, "providers", p.namespace, p.name)

	repo, err := git.GetRepo(repoURL, localPath)
	if err != nil {
		return nil, fmt.Errorf("failed to get repository: %w", err)
	}

	return repo.GetTagDate(ctx, version)
}

// WorktreeManager handles worktree lifecycle for a specific tag
type WorktreeManager struct {
	provider     *Provider
	tag          string
	repo         *git.Repo
	worktreePath string
	repoURL      string
}

// CreateWorktreeForTag creates a worktree for the specified tag and returns a manager
func (p *Provider) CreateWorktreeForTag(ctx context.Context, tag string) (*WorktreeManager, error) {
	tracer := otel.Tracer("opentofu-registry-backend")
	ctx, span := tracer.Start(ctx, "providers.create_worktree_for_tag")
	defer span.End()

	span.SetAttributes(
		attribute.String("provider.namespace", p.namespace),
		attribute.String("provider.name", p.name),
		attribute.String("provider.tag", tag),
	)

	repoURL := fmt.Sprintf("https://github.com/%s/terraform-provider-%s", p.namespace, p.name)
	localPath := filepath.Join(p.config.WorkDir, "providers", p.namespace, p.name)

	span.SetAttributes(
		attribute.String("git.repo_url", repoURL),
		attribute.String("git.local_path", localPath),
	)

	repo, err := git.GetRepo(repoURL, localPath)
	if err != nil {
		span.RecordError(err)
		return nil, fmt.Errorf("failed to get repository: %w", err)
	}

	if err := repo.Clone(ctx); err != nil {
		span.RecordError(err)
		return nil, fmt.Errorf("failed to clone repository: %w", err)
	}

	// Create worktree for the specific tag (add v prefix if not present)
	gitTag := tag
	if !strings.HasPrefix(tag, "v") {
		gitTag = "v" + tag
	}
	worktreePath := filepath.Join(localPath, "worktrees", tag)
	if err := repo.AddWorktree(ctx, gitTag, worktreePath); err != nil {
		span.RecordError(err)
		return nil, fmt.Errorf("failed to create worktree for tag %s: %w", gitTag, err)
	}

	return &WorktreeManager{
		provider:     p,
		tag:          tag,
		repo:         repo,
		worktreePath: worktreePath,
		repoURL:      repoURL,
	}, nil
}

// Cleanup removes the worktree
func (wm *WorktreeManager) Cleanup(ctx context.Context) error {
	tracer := otel.Tracer("opentofu-registry-backend")
	ctx, span := tracer.Start(ctx, "providers.cleanup_worktree")
	defer span.End()

	span.SetAttributes(
		attribute.String("provider.namespace", wm.provider.namespace),
		attribute.String("provider.name", wm.provider.name),
		attribute.String("provider.tag", wm.tag),
	)

	if err := wm.repo.RemoveWorktree(ctx, wm.tag); err != nil {
		span.RecordError(err)
		return fmt.Errorf("failed to remove worktree for tag %s: %w", wm.tag, err)
	}

	return nil
}

// CheckoutVersionForScraping creates a worktree for a tag and returns the directory path and cleanup function
func (p *Provider) CheckoutVersionForScraping(ctx context.Context, tag string) (string, func(), error) {
	wm, err := p.CreateWorktreeForTag(ctx, tag)
	if err != nil {
		return "", nil, err
	}

	cleanup := func() {
		if cleanupErr := wm.Cleanup(ctx); cleanupErr != nil {
			// Log the error but don't fail
			tracer := otel.Tracer("opentofu-registry-backend")
			_, span := tracer.Start(ctx, "providers.cleanup_error")
			span.RecordError(cleanupErr)
			span.End()
		}
	}

	return wm.worktreePath, cleanup, nil
}

// DetectLicensesInDirectory detects licenses in a given directory path
func (p *Provider) DetectLicensesInDirectory(ctx context.Context, directory string) (license.List, error) {
	tracer := otel.Tracer("opentofu-registry-backend")
	ctx, span := tracer.Start(ctx, "providers.detect_licenses_in_directory")
	defer span.End()

	span.SetAttributes(
		attribute.String("provider.namespace", p.namespace),
		attribute.String("provider.name", p.name),
		attribute.String("directory", directory),
	)

	// Get filesystem access from the provided directory
	fsys := os.DirFS(directory)
	readDirFS, ok := fsys.(fs.ReadDirFS)
	if !ok {
		err := fmt.Errorf("filesystem does not implement ReadDirFS")
		span.RecordError(err)
		return nil, err
	}

	// Create license detector
	detector, err := license.New(p.config.License)
	if err != nil {
		span.RecordError(err)
		return nil, fmt.Errorf("failed to create license detector: %w", err)
	}

	// Generate repo URL for context
	repoURL := fmt.Sprintf("https://github.com/%s/terraform-provider-%s", p.namespace, p.name)

	// Detect licenses in the directory
	licenses, err := detector.Detect(ctx, readDirFS, repoURL)
	if err != nil {
		span.RecordError(err)
		return nil, fmt.Errorf("failed to detect licenses in directory %s: %w", directory, err)
	}

	span.SetAttributes(attribute.Int("provider.licenses_count", len(licenses)))
	return licenses, nil
}
