package providertypes

import (
	"github.com/opentofu/registry-ui/internal/license"
)

// ProviderVersion describes a single provider version.
type ProviderVersion struct {
	ProviderVersionDescriptor

	Docs ProviderDocs `json:"docs"`

	CDKTFDocs map[CDKTFLanguage]ProviderDocs `json:"cdktf_docs"`

	Licenses license.List `json:"license"`

	// IncompatibleLicense indicates that there are no licenses or there is one or more license that are not approved.
	IncompatibleLicense bool `json:"incompatible_license"`

	Link string `json:"link"`
}

// ProviderDocs describes either a provider or a CDKTF language.
type ProviderDocs struct {
	Root        *ProviderDocItem  `json:"index,omitempty"`
	Resources   []ProviderDocItem `json:"resources"`
	DataSources []ProviderDocItem `json:"datasources"`
	Functions   []ProviderDocItem `json:"functions"`
	Guides      []ProviderDocItem `json:"guides"`
}

// ProviderDocItem describes a single documentation item.
type ProviderDocItem struct {
	Name     DocItemName `json:"name"`
	EditLink string      `json:"edit_link"`

	Title       string `json:"title"`
	Subcategory string `json:"subcategory"`
	Description string `json:"description"`
}
