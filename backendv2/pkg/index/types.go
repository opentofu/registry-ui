package index

import "time"

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
