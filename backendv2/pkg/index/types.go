package index

import (
	"time"

	"github.com/opentofu/registry-ui/pkg/license"
)

// IndexResponse contains the results of indexing a provider
type IndexResponse struct {
	TotalVersions     int
	ProcessedVersions int
	SkippedVersions   int
	FailedVersions    int
	Results           []VersionResult
}

// VersionResult contains the result of processing a single version
type VersionResult struct {
	Version         string
	Licenses        license.List
	Redistributable bool
	LicensesStr     string
	Explanation     string
	Error           error
	Duration        time.Duration
}

// MultiProviderIndexResponse contains the results of indexing multiple providers
type MultiProviderIndexResponse struct {
	TotalProviders     int
	ProcessedProviders int
	SkippedProviders   int
	FailedProviders    int
	ProviderResults    []ProviderIndexResult
}

// ProviderIndexResult contains the result of indexing a single provider
type ProviderIndexResult struct {
	Namespace string
	Name      string
	Response  *IndexResponse
	Error     error
	Duration  time.Duration
}
