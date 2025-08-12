package github

import (
	"context"
	"fmt"
	"log/slog"

	"github.com/google/go-github/v67/github"
	"golang.org/x/oauth2"

	"github.com/opentofu/registry-ui/pkg/config"
)

type Client struct {
	client *github.Client
	config *config.GitHubConfig
}

func NewClient(cfg *config.GitHubConfig) *Client {
	ts := oauth2.StaticTokenSource(
		&oauth2.Token{AccessToken: cfg.Token},
	)
	tc := oauth2.NewClient(context.Background(), ts)

	return &Client{
		client: github.NewClient(tc),
		config: cfg,
	}
}

func (c *Client) GetRepositoryMetadata(ctx context.Context, owner, name string) (*RepositoryMetadata, error) {
	slog.DebugContext(ctx, "Fetching repository metadata from GitHub", "owner", owner, "name", name)

	repo, _, err := c.client.Repositories.Get(ctx, owner, name)
	if err != nil {
		return nil, fmt.Errorf("failed to get repository %s/%s: %w", owner, name, err)
	}

	metadata := &RepositoryMetadata{
		Owner:       owner,
		Name:        name,
		Stars:       int64(repo.GetStargazersCount()),
		Forks:       int64(repo.GetForksCount()),
		Description: repo.GetDescription(),
		IsFork:      repo.GetFork(),
	}

	if repo.GetFork() && repo.GetParent() != nil {
		parent := repo.GetParent()
		metadata.ParentOwner = parent.GetOwner().GetLogin()
		metadata.ParentName = parent.GetName()
	}

	slog.DebugContext(ctx, "Retrieved repository metadata",
		"owner", owner, "name", name,
		"stars", metadata.Stars, "forks", metadata.Forks,
		"is_fork", metadata.IsFork)

	return metadata, nil
}
