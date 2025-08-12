package index

import (
	"time"
	
	"github.com/opentofu/registry-ui/pkg/license"
)

type ProviderAddr struct {
	Display   string `json:"display"`
	Namespace string `json:"namespace"`
	Name      string `json:"name"`
}

type ProviderVersionDescriptor struct {
	ID        string    `json:"id"`
	Published time.Time `json:"published"`
}

type ProviderIndexData struct {
	Addr           ProviderAddr                `json:"addr"`
	Description    string                      `json:"description"`
	Stars          int64                       `json:"popularity"`
	ForkCount      int64                       `json:"fork_count"`
	Versions       []ProviderVersionDescriptor `json:"versions"`
	IsBlocked      bool                        `json:"is_blocked"`
	CanonicalAddr  *ProviderAddr               `json:"canonical_addr,omitempty"`
	ForkOf         *ProviderAddr               `json:"fork_of,omitempty"`
	Link           string                      `json:"link,omitempty"`
	ReverseAliases []ProviderAddr              `json:"reverse_aliases,omitempty"`
	Warnings       []string                    `json:"warnings,omitempty"`
}

type DatabaseProviderData struct {
	Namespace      string
	Name           string
	Warnings       []string
	RepoOwner      string
	RepoName       string
	Stars          int64
	ForkCount      int64
	Description    string
	IsFork         bool
	ParentOwner    string
	ParentName     string
	Versions       []string
	VersionDates   []time.Time
	ReverseAliases []ProviderAddr
	CanonicalAddr  *ProviderAddr
	ForkOf         *ProviderAddr
}

// IndexResponse represents the response from indexing a provider version
type IndexResponse struct {
	TotalVersions     int            `json:"total_versions"`
	ProcessedVersions int            `json:"processed_versions"`
	SkippedVersions   int            `json:"skipped_versions"`
	FailedVersions    int            `json:"failed_versions"`
	Results           []VersionResult `json:"results"`
}

// VersionResult represents the result of processing a single version
type VersionResult struct {
	Version        string        `json:"version"`
	Success        bool          `json:"success"`
	Error          error         `json:"error,omitempty"`
	Duration       time.Duration `json:"duration"`
	Licenses       license.List  `json:"licenses,omitempty"`
	LicensesStr    string        `json:"licenses_str,omitempty"`
	Redistributable bool         `json:"redistributable"`
	Explanation    string        `json:"explanation,omitempty"`
}

// MultiProviderIndexResponse represents the response from indexing multiple providers
type MultiProviderIndexResponse struct {
	TotalProviders     int                   `json:"total_providers"`
	ProcessedProviders int                   `json:"processed_providers"`
	SkippedProviders   int                   `json:"skipped_providers"`
	FailedProviders    int                   `json:"failed_providers"`
	ProviderResults    []ProviderIndexResult `json:"provider_results"`
}

// ProviderIndexResult represents the result of processing a single provider
type ProviderIndexResult struct {
	Namespace string         `json:"namespace"`
	Name      string         `json:"name"`
	Response  *IndexResponse `json:"response,omitempty"`
	Error     error          `json:"error,omitempty"`
	Duration  time.Duration  `json:"duration"`
}
