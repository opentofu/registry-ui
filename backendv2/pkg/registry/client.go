package registry

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	"go.opentelemetry.io/otel/attribute"

	"github.com/opentofu/registry-ui/pkg/telemetry"
)

const registryURL = "https://github.com/opentofu/registry.git"

type Client struct {
	path string
}

func New(path string) (*Client, error) {
	if path == "" {
		return nil, fmt.Errorf("repo path cannot be empty")
	}

	absPath, err := filepath.Abs(path)
	if err != nil {
		return nil, fmt.Errorf("failed to get absolute path: %w", err)
	}

	return &Client{
		path: absPath,
	}, nil
}

func (r *Client) Clone(ctx context.Context) error {
	// If .git directory exists, assume it's already cloned
	if _, err := os.Stat(filepath.Join(r.path, ".git")); err == nil {
		return nil
	}

	// If directory exists but no .git, remove it and start fresh
	if _, err := os.Stat(r.path); err == nil {
		// Directory exists but no valid .git, remove it completely
		if err := os.RemoveAll(r.path); err != nil {
			return fmt.Errorf("failed to remove existing directory: %w", err)
		}
	}

	parentDir := filepath.Dir(r.path)
	if err := os.MkdirAll(parentDir, 0o755); err != nil {
		return fmt.Errorf("failed to create parent directory: %w", err)
	}

	cmd := exec.CommandContext(ctx, "git", "clone", "--depth", "1", registryURL, r.path)
	if output, err := cmd.CombinedOutput(); err != nil {
		return fmt.Errorf("failed to clone repository: %w\n%s", err, output)
	}

	return nil
}

// Update pulls the latest changes from the remote repository.
func (r *Client) Update(ctx context.Context) error {
	ctx, span := telemetry.Tracer().Start(ctx, "registry.update")
	defer span.End()

	gitDir := filepath.Join(r.path, ".git")
	if _, err := os.Stat(gitDir); os.IsNotExist(err) {
		return r.Clone(ctx)
	}

	resetCmd := exec.CommandContext(ctx, "git", "-C", r.path, "reset", "--hard", "HEAD")
	if err := resetCmd.Run(); err != nil {
		// If reset fails, remove and re-clone
		os.RemoveAll(r.path)
		return r.Clone(ctx)
	}
	span.AddEvent("git reset --hard HEAD completed")

	// Ensure we're on main branch
	checkoutCmd := exec.CommandContext(ctx, "git", "-C", r.path, "checkout", "main")
	if err := checkoutCmd.Run(); err != nil {
		// If checkout fails, remove and re-clone
		os.RemoveAll(r.path)
		return r.Clone(ctx)
	}
	span.AddEvent("git checkout main completed")

	// Now pull latest changes
	cmd := exec.CommandContext(ctx, "git", "-C", r.path, "pull", "--ff-only")
	if err := cmd.Run(); err != nil {
		// If pull fails, remove and re-clone
		os.RemoveAll(r.path)
		return r.Clone(ctx)
	}
	span.AddEvent("git pull --ff-only completed")

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

func readJSONFile(ctx context.Context, path string, dest any) error {
	_, span := telemetry.Tracer().Start(ctx, "registry.read-json-file")
	defer span.End()

	span.SetAttributes(attribute.String("file.path", path))

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
