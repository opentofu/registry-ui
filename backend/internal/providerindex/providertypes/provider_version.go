package providertypes

import (
	"github.com/opentofu/registry-ui/internal/license"
)

// ProviderVersion describes a single provider version.
//
// swagger:model ProviderVersion
type ProviderVersion struct {
	ProviderVersionDescriptor

	// required:true
	Docs ProviderDocs `json:"docs"`

	// required:true
	CDKTFDocs map[CDKTFLanguage]ProviderDocs `json:"cdktf_docs"`

	// required:true
	Licenses license.List `json:"license"`

	// IncompatibleLicense indicates that there are no licenses or there is one or more license that are not approved.
	// required:true
	IncompatibleLicense bool `json:"incompatible_license"`

	// required:false
	Link string `json:"link"`
}

// ProviderDocs describes either a provider or a CDKTF language.
//
// swagger:model ProviderDocs
type ProviderDocs struct {
	Root *ProviderDocItem `json:"index,omitempty"`
	// required: true
	Resources []ProviderDocItem `json:"resources"`
	// required: true
	DataSources []ProviderDocItem `json:"datasources"`
	// required: true
	Functions []ProviderDocItem `json:"functions"`
	// required: true
	Guides []ProviderDocItem `json:"guides"`
}

// ProviderDocItem describes a single documentation item.
//
// swagger:model ProviderDocItem
type ProviderDocItem struct {
	Name     DocItemName `json:"name"`
	EditLink string      `json:"edit_link"`

	Title       string `json:"title"`
	Subcategory string `json:"subcategory"`
	Description string `json:"description"`
}
