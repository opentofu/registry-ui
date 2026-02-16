package module

import "github.com/opentofu/registry-ui/pkg/license"

// Variable represents a module variable with its metadata
type Variable struct {
	Type        string      `json:"type"`
	Default     interface{} `json:"default,omitempty"` // Can be any JSON type
	Description string      `json:"description,omitempty"`
	Deprecated  string      `json:"deprecated,omitempty"`
	Sensitive   bool        `json:"sensitive,omitempty"`
	Required    bool        `json:"required"`
	Ephemeral   bool        `json:"ephemeral,omitempty"`
}

// Output represents a module output with its metadata
type Output struct {
	Sensitive   bool     `json:"sensitive,omitempty"`
	Ephemeral   bool     `json:"ephemeral,omitempty"`
	Deprecated  string   `json:"deprecated,omitempty"`
	DependsOn   []string `json:"dependsOn,omitempty"`
	Description string   `json:"description,omitempty"`
}

// Provider represents a provider configuration
type Provider struct {
	Alias             string `json:"alias,omitempty"`
	Name              string `json:"name"`
	FullName          string `json:"full_name"`
	VersionConstraint string `json:"version_constraint,omitempty"`
	ModuleAddress     string `json:"module_address,omitempty"`
}

// Resource represents a resource or data source
type Resource struct {
	Address string `json:"address"`
	Mode    string `json:"mode"`
	Type    string `json:"type"`
	Name    string `json:"name"`
}

// Dependency represents a module call dependency
type Dependency struct {
	Name              string `json:"name"`
	Source            string `json:"source"`
	VersionConstraint string `json:"version_constraint"`
}

// BaseComponentData contains fields common to all components (modules, submodules, examples)
type BaseComponentData struct {
	Variables   map[string]Variable `json:"variables"`
	Outputs     map[string]Output   `json:"outputs"`
	SchemaError string              `json:"schema_error"`
	Readme      bool                `json:"readme"`
	EditLink    string              `json:"edit_link"`
}

// ModuleComponentData extends BaseComponentData with module-specific fields
// Used by modules and submodules (but not examples)
type ModuleComponentData struct {
	BaseComponentData
	Providers    []Provider    `json:"providers"`
	Dependencies []Dependency `json:"dependencies"`
	Resources    []Resource    `json:"resources"`
}

// ExampleData represents an example's structure (only needs base fields)
type ExampleData struct {
	BaseComponentData
}

// SubmoduleData represents a submodule's structure
type SubmoduleData struct {
	ModuleComponentData
}

// ModuleData represents the complete module structure including metadata and nested components
type ModuleData struct {
	ModuleComponentData
	ID                  string                   `json:"id"`
	Published           string                   `json:"published"`
	Link                string                   `json:"link"`
	VCSRepository       string                   `json:"vcs_repository"`
	Licenses            []license.License        `json:"licenses"`
	IncompatibleLicense bool                     `json:"incompatible_license"`
	Submodules          map[string]SubmoduleData `json:"submodules,omitempty"`
	Examples            map[string]ExampleData   `json:"examples,omitempty"`
}
