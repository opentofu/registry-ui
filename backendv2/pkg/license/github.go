package license

import "strings"

// generateGitHubLink creates a GitHub blob URL for the given repository URL and file path
func generateGitHubLink(repoURL, filePath string) string {
	if repoURL == "" || filePath == "" {
		return ""
	}

	// Convert git URLs to GitHub web URLs
	// Handle both https://github.com/owner/repo.git and git@github.com:owner/repo.git formats
	var githubURL string
	if strings.HasPrefix(repoURL, "git@github.com:") {
		// git@github.com:owner/repo.git -> https://github.com/owner/repo
		githubURL = "https://github.com/" + strings.TrimSuffix(strings.TrimPrefix(repoURL, "git@github.com:"), ".git")
	} else if strings.Contains(repoURL, "github.com") {
		// https://github.com/owner/repo.git -> https://github.com/owner/repo
		githubURL = strings.TrimSuffix(repoURL, ".git")
	} else {
		return ""
	}

	// Create blob URL pointing to main branch
	return githubURL + "/blob/main/" + filePath
}
