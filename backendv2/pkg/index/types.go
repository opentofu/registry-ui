package index

import "time"

// ModuleVersionIndex represents the complete index structure for a module
// This matches the OpenTofu Registry API format exactly
type ModuleVersionIndex struct {
	Addr               ModuleAddr    `json:"addr"`
	Description        string        `json:"description,omitempty"`
	Versions           []VersionInfo `json:"versions"`
	IsBlocked          bool          `json:"is_blocked"`
	BlockedReason      *string       `json:"blocked_reason,omitempty"`
	Popularity         int           `json:"popularity"`             // from repository_stats.stars
	ForkCount          int           `json:"fork_count"`             // from repository_stats.forks
	ForkOf             *ModuleAddr   `json:"fork_of,omitempty"`      // if is_fork = true
	ForkOfLink         *string       `json:"fork_of_link,omitempty"` // GitHub URL to parent
	UpstreamPopularity int           `json:"upstream_popularity"`    // parent repo stars
	UpstreamForkCount  int           `json:"upstream_fork_count"`    // parent repo forks
}

// ModuleAddr represents a module address in the registry
type ModuleAddr struct {
	Display   string `json:"display"`
	Namespace string `json:"namespace"`
	Name      string `json:"name"`
	Target    string `json:"target"`
}

// VersionInfo represents version information for a module
type VersionInfo struct {
	ID        string     `json:"id"`                  // The version usually
	Published *time.Time `json:"published,omitempty"` // When the version was published
}

// ProviderVersionIndex represents the complete index structure for a provider
// This matches the OpenTofu Registry API format exactly
type ProviderVersionIndex struct {
	Addr               ProviderAddr  `json:"addr"`
	Description        string        `json:"description,omitempty"`
	Versions           []VersionInfo `json:"versions"`
	Warnings           []string      `json:"warnings,omitempty"` // from providers.warnings
	IsBlocked          bool          `json:"is_blocked"`
	BlockedReason      *string       `json:"blocked_reason,omitempty"`
	Popularity         int           `json:"popularity"`             // from repository_stats.stars
	ForkCount          int           `json:"fork_count"`             // from repository_stats.forks
	ForkOf             *ProviderAddr `json:"fork_of,omitempty"`      // if is_fork = true
	ForkOfLink         *string       `json:"fork_of_link,omitempty"` // GitHub URL to parent
	UpstreamPopularity int           `json:"upstream_popularity"`    // parent repo stars
	UpstreamForkCount  int           `json:"upstream_fork_count"`    // parent repo forks
}

// ProviderAddr represents a provider address in the registry
type ProviderAddr struct {
	Display   string `json:"display"`
	Namespace string `json:"namespace"`
	Name      string `json:"name"`
}

// GlobalModuleIndex represents the global module index file
type GlobalModuleIndex struct {
	Modules []ModuleEntry `json:"modules"`
}

// ModuleEntry represents a module entry in the global index
type ModuleEntry struct {
	Addr          ModuleAddr `json:"addr"`
	Description   string     `json:"description,omitempty"`
	LatestVersion string     `json:"latest_version"`
	PublishedAt   time.Time  `json:"published_at"`
}

// GlobalProviderIndex represents the global provider index file
type GlobalProviderIndex struct {
	Providers []ProviderEntry `json:"providers"`
}

// ProviderEntry represents a provider entry in the global index
type ProviderEntry struct {
	Addr          ProviderAddr  `json:"addr"`
	Description   string        `json:"description,omitempty"`
	LatestVersion string        `json:"latest_version"`
	PublishedAt   time.Time     `json:"published_at"`
	Link          string        `json:"link,omitempty"`         // GitHub repository URL
	Warnings      []string      `json:"warnings,omitempty"`     // Provider warnings
	Popularity    int           `json:"popularity"`             // Repository stars
	ForkCount     int           `json:"fork_count"`             // Repository fork count
	ForkOf        *ProviderAddr `json:"fork_of,omitempty"`      // Parent provider if this is a fork
	ForkOfLink    *string       `json:"fork_of_link,omitempty"` // GitHub URL to parent repo
}

// RepositoryStats holds repository statistics from GitHub
type RepositoryStats struct {
	Stars       int      `db:"stars"`
	Forks       int      `db:"forks"`
	Watchers    int      `db:"watchers"`
	Subscribers int      `db:"subscribers"`
	Topics      []string `db:"topics"`
}

// RepositoryMetadata holds repository metadata including fork information
type RepositoryMetadata struct {
	Description        *string `db:"description"`
	Homepage           *string `db:"homepage"`
	Language           *string `db:"language"`
	Archived           bool    `db:"archived"`
	DefaultBranch      *string `db:"default_branch"`
	IsFork             bool    `db:"is_fork"`
	ParentOrganisation *string `db:"parent_organisation"`
	ParentName         *string `db:"parent_name"`
}
