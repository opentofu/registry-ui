package license

import (
	"context"
	"fmt"
	"strings"

	"go.opentelemetry.io/otel/attribute"

	"github.com/opentofu/registry-ui/pkg/telemetry"
)

// generateGitHubLink creates a GitHub blob URL for the given repository URL and file path
func generateGitHubLink(repoURL, filePath string) string {
	if repoURL == "" || filePath == "" {
		return ""
	}

	// Convert git URLs to GitHub web URLs
	// Handle both https://github.com/owner/repo.git and git@github.com:owner/repo.git formats
	var githubURL string
	if after, ok := strings.CutPrefix(repoURL, "git@github.com:"); ok {
		// git@github.com:owner/repo.git -> https://github.com/owner/repo
		githubURL = "https://github.com/" + strings.TrimSuffix(after, ".git")
	} else if strings.Contains(repoURL, "github.com") {
		// https://github.com/owner/repo.git -> https://github.com/owner/repo
		githubURL = strings.TrimSuffix(repoURL, ".git")
	} else {
		return ""
	}

	// Create blob URL pointing to main branch
	return githubURL + "/blob/main/" + filePath
}

// detectLicenseFromGitHub detects license using GitHub API
func (d detector) detectLicenseFromGitHub(ctx context.Context, repoURL string) (*License, error) {
	ctx, span := telemetry.Tracer().Start(ctx, "detect-license-from-github")
	defer span.End()

	span.SetAttributes(attribute.String("repoURL", repoURL))
	if d.githubClient == nil {
		return nil, fmt.Errorf("github client not available")
	}

	spdxID, err := d.githubClient.DetectLicenseFromGitHub(ctx, repoURL)
	if err != nil {
		span.RecordError(err)
		return nil, fmt.Errorf("failed to detect license from GitHub: %w", err)
	}

	_, isCompatible := d.licenseMap[strings.ToLower(spdxID)]

	license := &License{
		SPDX:         spdxID,
		Confidence:   1.0, // GitHub API is authoritative when we're backing up
		IsCompatible: isCompatible,
		File:         "",
		Link:         repoURL,
	}

	return license, nil
}
