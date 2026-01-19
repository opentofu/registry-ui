package moduleschema

import (
	"github.com/zclconf/go-cty/cty"
)

type Schema struct {
	ProviderConfig map[string]ProviderConfigSchema `json:"provider_config,omitempty"`
	RootModule     ModuleSchema                    `json:"root_module"`
}

type ProviderConfigSchema struct {
	Name              string `json:"name"`
	FullName          string `json:"full_name"`
	VersionConstraint string `json:"version_constraint"`
}

type ModuleSchema struct {
	Resources      []Resource                      `json:"resources,omitempty"`
	ModuleCalls    map[string]ModuleCall           `json:"module_calls,omitempty"`
	ProviderConfig map[string]ProviderConfigSchema `json:"provider_config,omitempty"`
	Variables      map[string]Variable             `json:"variables,omitempty"`
	Outputs        map[string]Output               `json:"outputs,omitempty"`
}

type Resource struct {
	Address           string `json:"address"`
	Mode              string `json:"mode,omitempty"`
	Type              string `json:"type"`
	Name              string `json:"name"`
	ProviderConfigKey string `json:"provider_config_key"`
	SchemaVersion     int    `json:"schema_version"`
}

type ModuleCall struct {
	Module            ModuleSchema `json:"module"`
	Source            string       `json:"source"`
	VersionConstraint string       `json:"version_constraint"`
}

type Variable struct {
	Type        cty.Type `json:"type,omitempty"`
	Default     any      `json:"default,omitempty"`
	Description string   `json:"description,omitempty"`
	Sensitive   bool     `json:"sensitive,omitempty"`
	Required    bool     `json:"required,omitempty"`
}

type Output struct {
	Expression  Expression `json:"expression"`
	Description string     `json:"description,omitempty"`
	Sensitive   bool       `json:"sensitive,omitempty"`
}

type Expression struct {
	References []string `json:"references"`
}
