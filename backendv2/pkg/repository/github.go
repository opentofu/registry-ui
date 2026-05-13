package repository

import (
	"context"
	"fmt"
	"log/slog"
	"strings"

	"github.com/google/go-github/v84/github"
	"golang.org/x/oauth2"

	"github.com/opentofu/registry-ui/pkg/config"
	"github.com/opentofu/registry-ui/pkg/telemetry"
)

type Client struct {
	client *github.Client
	config *config.GitHubConfig
}

func NewClient(ctx context.Context, cfg *config.GitHubConfig) *Client {
	ts := oauth2.StaticTokenSource(
		&oauth2.Token{AccessToken: cfg.Token},
	)
	tc := oauth2.NewClient(ctx, ts)

	return &Client{
		client: github.NewClient(tc),
		config: cfg,
	}
}

func isRedirect(actualOwner, actualName, requestedOwner, requestedName string) bool {
	if !strings.EqualFold(actualOwner, requestedOwner) {
		return true
	}
	if !strings.EqualFold(actualName, requestedName) {
		return true
	}
	return false
}

func (c *Client) GetRepositoryMetadata(ctx context.Context, owner, name string) (*RepositoryMetadata, error) {
	ctx, span := telemetry.Tracer().Start(ctx, "repository.get_metadata")
	slog.DebugContext(ctx, "Fetching repository metadata from GitHub", "owner", owner, "name", name)
	defer span.End()

	repo, _, err := c.client.Repositories.Get(ctx, owner, name)
	if err != nil {
		return nil, fmt.Errorf("failed to get repository %s/%s: %w", owner, name, err)
	}

	// Get actual repository owner and name (may differ from requested if redirected)
	actualOwner := repo.GetOwner().GetLogin()
	actualName := repo.GetName()
	isRedirect := isRedirect(actualOwner, actualName, owner, name)
	metadata := &RepositoryMetadata{
		Owner:       owner,
		Name:        name,
		Stars:       int64(repo.GetStargazersCount()),
		Forks:       int64(repo.GetForksCount()),
		Description: repo.GetDescription(),
		IsFork:      repo.GetFork(),
		ActualOwner: actualOwner,
		ActualName:  actualName,
		IsRedirect:  isRedirect,
		OpenIssues:  int64(*repo.OpenIssuesCount),
		Subscribers: int64(*repo.SubscribersCount),
		Topics:      repo.Topics,
		// New metadata fields
		Homepage:        repo.GetHomepage(),
		Language:        repo.GetLanguage(),
		Archived:        repo.GetArchived(),
		DefaultBranch:   repo.GetDefaultBranch(),
		CreatedAtGitHub: repo.GetCreatedAt().Time,
		PushedAt:        repo.GetPushedAt().Time,
		UpdatedAtGitHub: repo.GetUpdatedAt().Time,
	}

	if repo.GetFork() && repo.GetParent() != nil {
		parent := repo.GetParent()
		metadata.ParentOwner = parent.GetOwner().GetLogin()
		metadata.ParentName = parent.GetName()
	}

	slog.DebugContext(ctx, "Retrieved repository metadata",
		"owner", owner, "name", name,
		"actual_owner", actualOwner, "actual_name", actualName,
		"is_redirect", isRedirect,
		"stars", metadata.Stars, "forks", metadata.Forks,
		"is_fork", metadata.IsFork)

	return metadata, nil
}

// GetRepositoryLicense gets the license information for a repository
func (c *Client) GetRepositoryLicense(ctx context.Context, owner, name string) (string, error) {
	slog.DebugContext(ctx, "Fetching repository license from GitHub", "owner", owner, "name", name)
	ctx, span := telemetry.Tracer().Start(ctx, "repository.get_license")
	defer span.End()

	// Get the repository license using GitHub's license detection
	license, _, err := c.client.Repositories.License(ctx, owner, name)
	if err != nil {
		span.RecordError(err)
		return "", fmt.Errorf("failed to get license for repository %s/%s: %w", owner, name, err)
	}

	if license == nil || license.License == nil {
		return "", fmt.Errorf("no license found for repository %s/%s", owner, name)
	}

	// Return the SPDX ID if available, otherwise the key
	spdxID := license.License.GetSPDXID()
	if spdxID != "" {
		return spdxID, nil
	}

	return license.License.GetKey(), nil
}

// parseGitHubURL extracts owner and repository name from a GitHub URL
func parseGitHubURL(gitHubURL string) (owner, repoName string) {
	// Remove https://github.com/ prefix
	if strings.HasPrefix(gitHubURL, "https://github.com/") {
		path := strings.TrimPrefix(gitHubURL, "https://github.com/")
		// Remove .git suffix if present
		path = strings.TrimSuffix(path, ".git")
		parts := strings.Split(path, "/")
		if len(parts) >= 2 {
			return parts[0], parts[1]
		}
	}

	// Handle git@github.com:owner/repo.git format
	if strings.HasPrefix(gitHubURL, "git@github.com:") {
		path := strings.TrimPrefix(gitHubURL, "git@github.com:")
		path = strings.TrimSuffix(path, ".git")
		parts := strings.Split(path, "/")
		if len(parts) >= 2 {
			return parts[0], parts[1]
		}
	}

	// Fallback: return empty strings if URL format is unexpected
	return "", ""
}

// DetectLicenseFromGitHub detects license using GitHub API and returns the SPDX identifier
func (c *Client) DetectLicenseFromGitHub(ctx context.Context, repoURL string) (string, error) {
	owner, repo := parseGitHubURL(repoURL)
	if owner == "" || repo == "" {
		return "", fmt.Errorf("could not parse GitHub URL: %s", repoURL)
	}

	slog.DebugContext(ctx, "Attempting GitHub API license detection",
		"owner", owner, "repo", repo, "url", repoURL)

	// Get repository license using GitHub's license detection API
	spdxID, err := c.GetRepositoryLicense(ctx, owner, repo)
	if err != nil {
		return "", fmt.Errorf("failed to get repository license: %w", err)
	}

	slog.InfoContext(ctx, "GitHub API detected license",
		"owner", owner, "repo", repo, "spdx_id", spdxID)

	return spdxID, nil
}
