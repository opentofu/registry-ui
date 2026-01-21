package moduleindex

import (
	"github.com/zclconf/go-cty/cty"

	"github.com/opentofu/registry-ui/internal/license"
)

type ModuleVersion struct {
	ModuleVersionDescriptor
	// Readme indicates that the version has a readme available.
	Details
	// Link holds the link to the repository browse URL.
	Link string `json:"link"`
	// VCSRepository holds the URL to the versioning system for this repository.
	VCSRepository string `json:"vcs_repository"`
	// Licenses is a list of licenses detected in the project.
	Licenses license.List `json:"licenses"`
	// IncompatibleLicense indicates that there are no licenses or there is one or more license that are not approved.
	IncompatibleLicense bool `json:"incompatible_license"`
	// Examples lists all examples for this version.
	Examples map[string]Example `json:"examples"`
	// Submodules lists all submodules of this version.
	Submodules map[string]Submodule `json:"submodules"`
}

// BaseDetails is an embedded struct describing a module or a submodule.
type BaseDetails struct {
	// Readme indicates that the submodule has a readme available.
	Readme    bool                `json:"readme"`
	EditLink  string              `json:"edit_link"`
	Variables map[string]Variable `json:"variables"`
	Outputs   map[string]Output   `json:"outputs"`

	// SchemaError contains an error message to show why the schema is not available. This should be shown to the user
	// as a warning message.
	SchemaError string `json:"schema_error"`
}

// Details is an embedded struct describing a module or a submodule.
type Details struct {
	BaseDetails
	Providers    []ProviderDependency `json:"providers"`
	Dependencies []ModuleDependency   `json:"dependencies"`
	Resources    []Resource           `json:"resources"`
}

// ModuleDependency describes a module call as a dependency as the UI expects it.
type ModuleDependency struct {
	Name              string `json:"name"`
	VersionConstraint string `json:"version_constraint"`
	Source            string `json:"source"`
}

// ProviderDependency describes a provider dependency of a module.
type ProviderDependency struct {
	Alias             string `json:"alias"`
	Name              string `json:"name"`
	FullName          string `json:"full_name"`
	VersionConstraint string `json:"version_constraint"`
}

// Variable describes a variable as the UI expects it.
type Variable struct {
	Type        cty.Type `json:"type"`
	Default     any      `json:"default"`
	Description string   `json:"description"`
	Sensitive   bool     `json:"sensitive"`
	Required    bool     `json:"required"`
}

// Output describes a module output as the UI expects it.
type Output struct {
	Sensitive   bool   `json:"sensitive"`
	Description string `json:"description"`
}

// Resource describes a resource a module uses as the UI expects it.
type Resource struct {
	Address string `json:"address"`
	Type    string `json:"type"`
	Name    string `json:"name"`
}

// Submodule describes a submodule within a module.
type Submodule struct {
	Details
}

// Example describes a single example for a Documentation. You can query the files and the readme using the
// corresponding API.
type Example struct {
	BaseDetails
}
