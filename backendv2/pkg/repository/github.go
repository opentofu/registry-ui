package repository

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"log/slog"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/google/go-github/v84/github"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/trace"
	"golang.org/x/oauth2"

	"github.com/opentofu/registry-ui/pkg/config"
	"github.com/opentofu/registry-ui/pkg/telemetry"
)

const githubGraphQLEndpoint = "https://api.github.com/graphql"

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

// buildRepositoryStatsQuery constructs a single GraphQL query that fetches
// stats for many repositories at once using aliases (r0, r1, ...). The whole
// query should be cheap regardless of batch size because it contains
// no paginated connections.
func buildRepositoryStatsQuery(repos []RepoIdentifier) string {
	var b strings.Builder
	b.WriteString("query {\n")
	for i, r := range repos {
		// GitHub caps repositories at 20 topics, so first: 20 captures all of them. If something has more than 20 topics, I think its fine to ignore.
		fmt.Fprintf(&b,
			"  r%d: repository(owner: %q, name: %q) { stargazerCount forkCount watchers { totalCount } issues(states: OPEN) { totalCount } repositoryTopics(first: 20) { nodes { topic { name } } } }\n",
			i, r.Owner, r.Name)
	}
	b.WriteString("  rateLimit { cost remaining resetAt }\n}")
	return b.String()
}

// GetRepositoryStatsBatch fetches stars, forks, watchers, and open-issue counts
// for a batch of repositories in a single GraphQL request.
func (c *Client) GetRepositoryStatsBatch(ctx context.Context, repos []RepoIdentifier, maxAttempts int) (map[RepoIdentifier]RepositoryStats, GraphQLRateLimit, error) {
	ctx, span := telemetry.Tracer().Start(ctx, "repository.get_stats_batch")
	defer span.End()
	span.SetAttributes(attribute.Int("repository.batch_size", len(repos)))

	result := make(map[RepoIdentifier]RepositoryStats, len(repos))
	if len(repos) == 0 {
		return result, GraphQLRateLimit{}, nil
	}

	reqBody, err := json.Marshal(map[string]string{"query": buildRepositoryStatsQuery(repos)})
	if err != nil {
		span.RecordError(err)
		return nil, GraphQLRateLimit{}, fmt.Errorf("failed to marshal graphql query: %w", err)
	}

	httpResp, err := c.doGraphQLWithRetry(ctx, span, reqBody, maxAttempts)
	if err != nil {
		return nil, GraphQLRateLimit{}, err
	}
	defer httpResp.Body.Close()

	if httpResp.StatusCode != http.StatusOK {
		err := fmt.Errorf("graphql request returned status %d", httpResp.StatusCode)
		span.RecordError(err)
		span.SetAttributes(attribute.Int("http.status_code", httpResp.StatusCode))
		return nil, GraphQLRateLimit{}, err
	}

	var gqlResp struct {
		Data   map[string]json.RawMessage `json:"data"`
		Errors []struct {
			Message string            `json:"message"`
			Type    string            `json:"type"`
			Path    []json.RawMessage `json:"path"`
		} `json:"errors"`
	}
	if err := json.NewDecoder(httpResp.Body).Decode(&gqlResp); err != nil {
		span.RecordError(err)
		return nil, GraphQLRateLimit{}, fmt.Errorf("failed to decode graphql response: %w", err)
	}

	var rl GraphQLRateLimit
	if raw, ok := gqlResp.Data["rateLimit"]; ok {
		err = json.Unmarshal(raw, &rl)
		if err != nil {
			span.RecordError(err)
			return nil, GraphQLRateLimit{}, fmt.Errorf("failed to unmarshal rate limit from graphql response: %w", err)
		}
	}
	span.SetAttributes(
		attribute.Int("graphql.rate_limit.cost", rl.Cost),
		attribute.Int("graphql.rate_limit.remaining", rl.Remaining),
	)

	if len(gqlResp.Errors) > 0 {
		span.SetAttributes(attribute.Int("graphql.errors", len(gqlResp.Errors)))
		slog.WarnContext(ctx, "Some repositories failed in GraphQL stats batch",
			"error_count", len(gqlResp.Errors),
			"first_error", gqlResp.Errors[0].Message)
	}

	type statsNode struct {
		StargazerCount int64 `json:"stargazerCount"`
		ForkCount      int64 `json:"forkCount"`
		Watchers       struct {
			TotalCount int64 `json:"totalCount"`
		} `json:"watchers"`
		Issues struct {
			TotalCount int64 `json:"totalCount"`
		} `json:"issues"`
		RepositoryTopics struct {
			Nodes []struct {
				Topic struct {
					Name string `json:"name"`
				} `json:"topic"`
			} `json:"nodes"`
		} `json:"repositoryTopics"`
	}

	for i, r := range repos {
		raw, ok := gqlResp.Data[fmt.Sprintf("r%d", i)]
		if !ok || string(raw) == "null" {
			// Repository failed individually (see errors array above) - skip it.
			continue
		}
		var node statsNode
		if err := json.Unmarshal(raw, &node); err != nil {
			slog.WarnContext(ctx, "Failed to decode repository stats node",
				"repository", fmt.Sprintf("%s/%s", r.Owner, r.Name), "error", err)
			continue
		}
		topics := make([]string, 0, len(node.RepositoryTopics.Nodes))
		for _, n := range node.RepositoryTopics.Nodes {
			topics = append(topics, n.Topic.Name)
		}
		result[r] = RepositoryStats{
			Stars:      node.StargazerCount,
			Forks:      node.ForkCount,
			Watchers:   node.Watchers.TotalCount,
			OpenIssues: node.Issues.TotalCount,
			Topics:     topics,
		}
	}

	span.SetAttributes(attribute.Int("repository.stats_fetched", len(result)))
	return result, rl, nil
}

// doGraphQLWithRetry POSTs the query body, retrying with backoff on GitHub
// secondary rate limits (403/429). It returns the first response that is not a
// secondary rate limit, or an error once attempts are exhausted.
func (c *Client) doGraphQLWithRetry(ctx context.Context, span trace.Span, reqBody []byte, maxAttempts int) (*http.Response, error) {
	for attempt := 1; ; attempt++ {
		resp, err := c.doGraphQL(ctx, reqBody)
		if err != nil {
			span.RecordError(err)
			return nil, err
		}

		// Any non-rate-limit status is the caller's to interpret.
		if resp.StatusCode != http.StatusForbidden && resp.StatusCode != http.StatusTooManyRequests {
			return resp, nil
		}

		wait := backoffDuration(getRetryAfter(resp.Header), attempt)
		resp.Body.Close()

		if attempt >= maxAttempts {
			err := fmt.Errorf("graphql secondary rate limited (status %d) after %d attempts", resp.StatusCode, attempt)
			span.RecordError(err)
			return nil, err
		}

		span.AddEvent("graphql secondary rate limit backoff")
		slog.WarnContext(ctx, "GraphQL secondary rate limit, backing off",
			"status", resp.StatusCode, "attempt", attempt, "wait", wait.String())

		select {
		case <-ctx.Done():
			return nil, ctx.Err()
		case <-time.After(wait):
		}
	}
}

// doGraphQL performs a single GraphQL POST request.
func (c *Client) doGraphQL(ctx context.Context, reqBody []byte) (*http.Response, error) {
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, githubGraphQLEndpoint, bytes.NewReader(reqBody))
	if err != nil {
		return nil, fmt.Errorf("failed to build graphql request: %w", err)
	}
	req.Header.Set("Content-Type", "application/json")

	resp, err := c.client.Client().Do(req)
	if err != nil {
		return nil, fmt.Errorf("graphql request failed: %w", err)
	}
	return resp, nil
}

// backoffDuration picks how long to wait before the next retry, preferring the
// server's Retry-After hint and falling back to quadratic backoff.
func backoffDuration(retryAfter time.Duration, attempt int) time.Duration {
	if retryAfter > 0 {
		return retryAfter
	}
	return time.Duration(attempt*attempt) * time.Second
}

func getRetryAfter(h http.Header) time.Duration {
	v := h.Get("Retry-After")
	if v == "" {
		return 0
	}
	if secs, err := strconv.Atoi(v); err == nil {
		return time.Duration(secs) * time.Second
	}
	if t, err := http.ParseTime(v); err == nil {
		return time.Until(t)
	}
	return 0
}
