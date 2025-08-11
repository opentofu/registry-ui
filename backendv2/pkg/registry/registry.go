package registry

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
)

const registryURL = "https://github.com/opentofu/registry.git"

type Registry struct {
	path string
}

func New(path string) (*Registry, error) {
	if path == "" {
		return nil, fmt.Errorf("repo path cannot be empty")
	}

	absPath, err := filepath.Abs(path)
	if err != nil {
		return nil, fmt.Errorf("failed to get absolute path: %w", err)
	}

	return &Registry{
		path: absPath,
	}, nil
}

func (r *Registry) Clone(ctx context.Context) error {
	if _, err := os.Stat(filepath.Join(r.path, ".git")); err == nil {
		return nil
	}

	parentDir := filepath.Dir(r.path)
	if err := os.MkdirAll(parentDir, 0755); err != nil {
		return fmt.Errorf("failed to create parent directory: %w", err)
	}

	cmd := exec.CommandContext(ctx, "git", "clone", "--depth", "1", registryURL, r.path)
	if output, err := cmd.CombinedOutput(); err != nil {
		return fmt.Errorf("failed to clone repository: %w\n%s", err, output)
	}

	return nil
}

func (r *Registry) Update(ctx context.Context) error {
	gitDir := filepath.Join(r.path, ".git")
	if _, err := os.Stat(gitDir); os.IsNotExist(err) {
		return r.Clone(ctx)
	}

	cmd := exec.CommandContext(ctx, "git", "-C", r.path, "pull", "--ff-only")
	if output, err := cmd.CombinedOutput(); err != nil {
		return fmt.Errorf("failed to update repository: %w\n%s", err, output)
	}

	return nil
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
		if strings.HasPrefix(pattern, "*") {
			return strings.HasSuffix(value, strings.TrimPrefix(pattern, "*"))
		}
		if strings.HasSuffix(pattern, "*") {
			return strings.HasPrefix(value, strings.TrimSuffix(pattern, "*"))
		}
	}

	return pattern == value
}

func readJSONFile(path string, v interface{}) error {
	data, err := os.ReadFile(path)
	if err != nil {
		return err
	}
	return json.Unmarshal(data, v)
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
