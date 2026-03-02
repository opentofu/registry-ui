package registry

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"strings"

	"github.com/opentofu/registry-ui/pkg/git"
)

const registryURL = "https://github.com/opentofu/registry.git"

type Client struct {
	repo *git.Repo
	path string
}

func New(path string) (*Client, error) {
	if path == "" {
		return nil, fmt.Errorf("repo path cannot be empty")
	}

	repo, err := git.GetRepo(registryURL, path)
	if err != nil {
		return nil, fmt.Errorf("failed to create git repo: %w", err)
	}

	return &Client{
		repo: repo,
		path: repo.LocalPath,
	}, nil
}

// Update clones the registry repository if needed and pulls the latest changes.
func (r *Client) Update(ctx context.Context) error {
	return r.repo.Update(ctx)
}

func parseFilter(filter string) []string {
	if filter == "" || filter == "*" {
		return nil
	}
	return strings.Split(filter, "/")
}

func matchesFilter(parts []string, filterParts []string) bool {
	if len(filterParts) == 0 {
		return true
	}

	if len(filterParts) > len(parts) {
		return false
	}

	for i, filterPart := range filterParts {
		if filterPart == "*" {
			continue
		}
		if !matchPattern(filterPart, parts[i]) {
			return false
		}
	}

	return true
}

func matchPattern(pattern, value string) bool {
	if pattern == "*" {
		return true
	}

	if strings.Contains(pattern, "*") {
		if strings.HasPrefix(pattern, "*") && strings.HasSuffix(pattern, "*") {
			return strings.Contains(value, strings.Trim(pattern, "*"))
		}
		trimmed, ok := strings.CutSuffix(pattern, "*")
		if ok {
			return strings.HasPrefix(value, trimmed)
		}
		if strings.HasSuffix(pattern, "*") {
			return strings.HasPrefix(value, strings.TrimSuffix(pattern, "*"))
		}
	}

	return pattern == value
}

func readJSONFile(_ context.Context, path string, dest any) error {
	data, err := os.ReadFile(path)
	if err != nil {
		return err
	}
	return json.Unmarshal(data, dest)
}

func fileExists(path string) bool {
	_, err := os.Stat(path)
	return err == nil
}

func listDirectories(path string) ([]string, error) {
	entries, err := os.ReadDir(path)
	if err != nil {
		return nil, err
	}

	var dirs []string
	for _, entry := range entries {
		if entry.IsDir() && !strings.HasPrefix(entry.Name(), ".") {
			dirs = append(dirs, entry.Name())
		}
	}
	return dirs, nil
}

func listJSONFiles(path string) ([]string, error) {
	entries, err := os.ReadDir(path)
	if err != nil {
		return nil, err
	}

	var files []string
	for _, entry := range entries {
		if !entry.IsDir() && strings.HasSuffix(entry.Name(), ".json") {
			name := strings.TrimSuffix(entry.Name(), ".json")
			files = append(files, name)
		}
	}
	return files, nil
}
