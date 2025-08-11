package providers

import (
	"time"

	"github.com/opentofu/registry-ui/pkg/license"
)

// ProviderVersion represents the JSON structure for index.json matching OpenTofu API spec
type ProviderVersion struct {
	ID                  string                        `json:"id"`
	Published           time.Time                     `json:"published"`
	Docs                ProviderDocs                  `json:"docs"`
	CDKTFDocs           map[string]ProviderDocs       `json:"cdktf_docs"`
	License             license.List                  `json:"license"`
	IncompatibleLicense bool                          `json:"incompatible_license"`
	Link                string                        `json:"link,omitempty"`
}

// ProviderDocs describes documentation for a provider or CDKTF language
type ProviderDocs struct {
	Index       *DocItem   `json:"index,omitempty"`
	Resources   []DocItem  `json:"resources,omitempty"`
	DataSources []DocItem  `json:"datasources,omitempty"`
	Functions   []DocItem  `json:"functions,omitempty"`
	Guides      []DocItem  `json:"guides,omitempty"`
	Ephemeral   []DocItem  `json:"ephemeral,omitempty"`
}

// DocItem describes a single documentation item
type DocItem struct {
	Name        string `json:"name"`
	Title       string `json:"title"`
	Subcategory string `json:"subcategory,omitempty"`
	Description string `json:"description,omitempty"`
	EditLink    string `json:"edit_link,omitempty"`
}