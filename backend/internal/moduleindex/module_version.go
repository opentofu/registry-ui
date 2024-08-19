package moduleindex

import (
	"github.com/opentofu/registry-ui/internal/license"
)

// swagger:model
type ModuleVersion struct {
	ModuleVersionDescriptor
	// Readme indicates that the version has a readme available.
	Details
	// Link holds the link to the repository browse URL.
	// required: false
	Link string `json:"link"`
	// IncompatibleLicense indicates that there are no licenses or there is one or more license that are not OSI
	// approved.
	// required:true
	IncompatibleLicense bool `json:"incompatible_license"`
	// VCSRepository holds the URL to the versioning system for this repository.
	// required:true
	VCSRepository string `json:"vcs_repository"`
	// Licenses is a list of licenses detected in the project.
	// required:true
	Licenses license.List `json:"licenses"`
	// Examples lists all examples for this version.
	// required:true
	Examples map[string]Example `json:"examples"`
	// Submodules lists all submodules of this version.
	// required:true
	Submodules map[string]Submodule `json:"submodules"`
}

// BaseDetails is an embedded struct describing a module or a submodule.
type BaseDetails struct {
	// Readme indicates that the submodule has a readme available.
	// required:true
	Readme bool `json:"readme"`
	// required:false
	EditLink string `json:"edit_link"`
	// required:true
	Variables map[string]Variable `json:"variables"`
	// required:true
	Outputs map[string]Output `json:"outputs"`

	// SchemaError contains an error message to show why the schema is not available. This should be shown to the user
	// as a warning message.
	// required:true
	SchemaError string `json:"schema_error"`
}

// Details is an embedded struct describing a module or a submodule.
//
// swagger:model ModuleDetails
type Details struct {
	BaseDetails
	// required:true
	Providers []ProviderDependency `json:"providers"`
	// required:true
	Dependencies []ModuleDependency `json:"dependencies"`
	// required:true
	Resources []Resource `json:"resources"`
}

// ModuleDependency describes a module call as a dependency as the UI expects it.
// swagger:model
type ModuleDependency struct {
	// required:true
	Name string `json:"name"`
	// required:true
	VersionConstraint string `json:"version_constraint"`
	// required:true
	Source string `json:"source"`
}

// ProviderDependency describes a provider dependency of a module.
// swagger:model
type ProviderDependency struct {
	// required:true
	Alias string `json:"alias"`
	// required:true
	Name string `json:"name"`
	// required:true
	FullName string `json:"full_name"`
	// required:true
	VersionConstraint string `json:"version_constraint"`
}

// Variable describes a variable as the UI expects it.
// swagger:model
type Variable struct {
	// required:true
	Type string `json:"type"`
	// required:true
	Default any `json:"default"`
	// required:true
	Description string `json:"description"`
	// required:true
	Sensitive bool `json:"sensitive"`
	// required:true
	Required bool `json:"required"`
}

// Output describes a module output as the UI expects it.
// swagger:model
type Output struct {
	// required:true
	Sensitive bool `json:"sensitive"`
	// required:true
	Description string `json:"description"`
}

// Resource describes a resource a module uses as the UI expects it.
// swagger:model
type Resource struct {
	// required:true
	Address string `json:"address"`
	// required:true
	Type string `json:"type"`
	// required:true
	Name string `json:"name"`
}

// Submodule describes a submodule within a module.
// swagger:model
type Submodule struct {
	Details
}

// Example describes a single example for a Documentation. You can query the files and the readme using the
// corresponding API.
// swagger:model ModuleExample
type Example struct {
	BaseDetails
}
