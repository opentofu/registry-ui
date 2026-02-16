package tofu

import "encoding/json"

// Config represents the complete configuration source
type Config struct {
	ProviderConfigs map[string]providerConfig `json:"provider_config,omitempty"`
	RootModule      module                    `json:"root_module"`
}

// ProviderConfig describes all of the provider configurations throughout the
// configuration tree, flattened into a single map for convenience since
// provider configurations are the one concept in OpenTofu that can span across
// module boundaries.
type providerConfig struct {
	Name              string         `json:"name,omitempty"`
	FullName          string         `json:"full_name,omitempty"`
	Alias             string         `json:"alias,omitempty"`
	VersionConstraint string         `json:"version_constraint,omitempty"`
	ModuleAddress     string         `json:"module_address,omitempty"`
	Expressions       map[string]any `json:"expressions,omitempty"`
}

type module struct {
	Outputs map[string]output `json:"outputs,omitempty"`
	// Resources are sorted in a user-friendly order that is undefined at this
	// time, but consistent.
	Resources   []resource            `json:"resources,omitempty"`
	ModuleCalls map[string]moduleCall `json:"module_calls,omitempty"`
	Variables   variables             `json:"variables,omitempty"`
}

type moduleCall struct {
	Source            string         `json:"source,omitempty"`
	Expressions       map[string]any `json:"expressions,omitempty"`
	CountExpression   *expression    `json:"count_expression,omitempty"`
	ForEachExpression *expression    `json:"for_each_expression,omitempty"`
	Module            *module        `json:"module,omitempty"`
	VersionConstraint string         `json:"version_constraint,omitempty"`
	DependsOn         []string       `json:"depends_on,omitempty"`
}

// variables is the JSON representation of the variables provided to the current
// plan.
type variables map[string]*variable

type variable struct {
	Type        json.RawMessage `json:"type,omitempty"`
	Default     json.RawMessage `json:"default,omitempty"`
	Description string          `json:"description,omitempty"`
	Required    bool            `json:"required,omitempty"`
	Sensitive   bool            `json:"sensitive,omitempty"`
	Ephemeral   bool            `json:"ephemeral,omitempty"`
	Deprecated  string          `json:"deprecated,omitempty"`
}

// Resource is the representation of a resource in the config
type resource struct {
	// Address is the absolute resource address
	Address string `json:"address,omitempty"`

	// Mode can be "managed" or "data"
	Mode string `json:"mode,omitempty"`

	Type string `json:"type,omitempty"`
	Name string `json:"name,omitempty"`

	// ProviderConfigKey is the key into "provider_configs" (shown above) for
	// the provider configuration that this resource is associated with.
	//
	// NOTE: If a given resource is in a ModuleCall, and the provider was
	// configured outside of the module (in a higher level configuration file),
	// the ProviderConfigKey will not match a key in the ProviderConfigs map.
	ProviderConfigKey string `json:"provider_config_key,omitempty"`

	// Provisioners is an optional field which describes any provisioners.
	// Connection info will not be included here.
	Provisioners []provisioner `json:"provisioners,omitempty"`

	// Expressions" describes the resource-type-specific  content of the
	// configuration block.
	Expressions map[string]any `json:"expressions,omitempty"`

	// SchemaVersion indicates which version of the resource type schema the
	// "values" property conforms to.
	SchemaVersion *uint64 `json:"schema_version,omitempty"`

	// CountExpression and ForEachExpression describe the expressions given for
	// the corresponding meta-arguments in the resource configuration block.
	// These are omitted if the corresponding argument isn't set.
	CountExpression   *expression `json:"count_expression,omitempty"`
	ForEachExpression *expression `json:"for_each_expression,omitempty"`

	DependsOn []string `json:"depends_on,omitempty"`
}

type output struct {
	Sensitive   bool        `json:"sensitive,omitempty"`
	Ephemeral   bool        `json:"ephemeral,omitempty"`
	Deprecated  string      `json:"deprecated,omitempty"`
	Expression  *expression `json:"expression,omitempty"`
	DependsOn   []string    `json:"depends_on,omitempty"`
	Description string      `json:"description,omitempty"`
}

type provisioner struct {
	Type        string         `json:"type,omitempty"`
	Expressions map[string]any `json:"expressions,omitempty"`
}

// expression represents any unparsed expression
type expression struct {
	// "constant_value" is set only if the expression contains no references to
	// other objects, in which case it gives the resulting constant value. This
	// is mapped as for the individual values in the common value
	// representation.
	ConstantValue json.RawMessage `json:"constant_value,omitempty"`

	// Alternatively, "references" will be set to a list of references in the
	// expression. Multi-step references will be unwrapped and duplicated for
	// each significant traversal step, allowing callers to more easily
	// recognize the objects they care about without attempting to parse the
	// expressions. Callers should only use string equality checks here, since
	// the syntax may be extended in future releases.
	References []string `json:"references,omitempty"`
}
