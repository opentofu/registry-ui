package module

import (
	"context"
	"fmt"
	"log/slog"
	"os"
	"path/filepath"
	"time"

	"github.com/opentofu/registry-ui/pkg/license"
	"github.com/opentofu/registry-ui/pkg/tofu"
)

// Parser handles parsing and transformation of module data
type Parser struct {
	workDir     string
	namespace   string
	name        string
	target      string
	version     string
	publishedAt *time.Time
}

// NewModuleParser creates a new module parser
func NewModuleParser(workDir, namespace, name, target, version string, publishedAt *time.Time) *Parser {
	return &Parser{
		workDir:     workDir,
		namespace:   namespace,
		name:        name,
		target:      target,
		version:     version,
		publishedAt: publishedAt,
	}
}

// BuildCompleteModuleStructure creates the complete module structure required by the registry API.
// Submodules and examples should be pre-collected in parallel before calling this function.
func (p *Parser) BuildCompleteModuleStructure(ctx context.Context, rootDir string, rootModuleData *tofu.Config, submodules map[string]SubmoduleData, examples map[string]ExampleData, licenses []license.License) (ModuleData, error) {
	slog.DebugContext(ctx, "Building complete module structure",
		"module", fmt.Sprintf("%s/%s/%s", p.namespace, p.name, p.target),
		"version", p.version)

	// Transform root module data
	rootTransformed, err := p.transformTofuShowOutput(rootModuleData)
	if err != nil {
		return ModuleData{}, fmt.Errorf("failed to transform root module data: %w", err)
	}

	// Set the edit link and readme status for the root module
	rootTransformed.EditLink = p.buildEditLink()
	rootTransformed.Readme = hasReadme(p.workDir)

	// Populate license links
	licensesWithLinks := make([]license.License, len(licenses))
	for i, lic := range licenses {
		lic.Link = p.buildLicenseLink(lic.File)
		licensesWithLinks[i] = lic
	}

	// Build the module structure
	moduleData := ModuleData{
		ModuleComponentData: rootTransformed,
		ID:                  p.version,
		Published:           p.formatPublishedDate(),
		Link:                p.buildRepoLink(),
		VCSRepository:       p.buildVCSRepository(),
		Licenses:            licensesWithLinks,
		IncompatibleLicense: p.hasIncompatibleLicense(licenses),
		Submodules:          submodules,
		Examples:            examples,
	}

	slog.DebugContext(ctx, "Successfully built complete module structure",
		"submodules_count", len(submodules),
		"examples_count", len(examples),
		"licenses_count", len(licenses))

	return moduleData, nil
}

// BuildSubmoduleData transforms submodule tofu config into storage format
func (p *Parser) BuildSubmoduleData(ctx context.Context, submoduleName string, tofuConfig *tofu.Config) (SubmoduleData, error) {
	slog.DebugContext(ctx, "Building submodule data structure",
		"module", fmt.Sprintf("%s/%s/%s", p.namespace, p.name, p.target),
		"version", p.version,
		"submodule", submoduleName)

	// Transform the raw tofu config
	transformed, err := p.transformTofuShowOutput(tofuConfig)
	if err != nil {
		return SubmoduleData{}, fmt.Errorf("failed to transform submodule data: %w", err)
	}

	// Set the edit link and readme status (transformTofuShowOutput leaves them empty/false)
	transformed.EditLink = p.buildSubmoduleEditLink(submoduleName)
	transformed.Readme = hasReadme(filepath.Join(p.workDir, "modules", submoduleName))

	slog.DebugContext(ctx, "Successfully built submodule data structure", "submodule", submoduleName)
	return SubmoduleData{ModuleComponentData: transformed}, nil
}

// BuildExampleData transforms example tofu config into storage format
func (p *Parser) BuildExampleData(ctx context.Context, exampleName string, tofuConfig *tofu.Config) (ExampleData, error) {
	slog.DebugContext(ctx, "Building example data structure",
		"module", fmt.Sprintf("%s/%s/%s", p.namespace, p.name, p.target),
		"version", p.version,
		"example", exampleName)

	// Transform the raw tofu config
	transformed, err := p.transformTofuShowOutput(tofuConfig)
	if err != nil {
		return ExampleData{}, fmt.Errorf("failed to transform example data: %w", err)
	}

	// Build the example structure (examples only need base component data, not providers/resources)
	exampleData := ExampleData{
		BaseComponentData: BaseComponentData{
			Variables:   transformed.Variables,
			Outputs:     transformed.Outputs,
			SchemaError: "",
			Readme:      hasReadme(filepath.Join(p.workDir, "examples", exampleName)),
			EditLink:    p.buildExampleEditLink(exampleName),
		},
	}

	slog.DebugContext(ctx, "Successfully built example data structure", "example", exampleName)
	return exampleData, nil
}

// transformTofuShowOutput transforms the raw tofu show -json output to registry format
func (p *Parser) transformTofuShowOutput(tofuData *tofu.Config) (ModuleComponentData, error) {
	// Transform variables
	transformedVars := make(map[string]Variable)
	for varName, varData := range tofuData.RootModule.Variables {
		transformedVars[varName] = Variable{
			Type:        string(varData.Type), // Convert json.RawMessage to string
			Default:     varData.Default,
			Description: varData.Description,
			Deprecated:  varData.Deprecated,
			Sensitive:   varData.Sensitive,
			Required:    varData.Required,
			Ephemeral:   varData.Ephemeral,
		}
	}

	// Transform outputs
	transformedOutputs := make(map[string]Output)
	for outputName, outputData := range tofuData.RootModule.Outputs {
		transformedOutputs[outputName] = Output{
			Sensitive:   outputData.Sensitive,
			Ephemeral:   outputData.Ephemeral,
			Deprecated:  outputData.Deprecated,
			DependsOn:   outputData.DependsOn,
			Description: outputData.Description,
		}
	}

	// Transform providers
	var providers []Provider
	for _, providerData := range tofuData.ProviderConfigs {
		providers = append(providers, Provider{
			Alias:             providerData.Alias,
			Name:              providerData.Name,
			FullName:          providerData.FullName,
			VersionConstraint: providerData.VersionConstraint,
			ModuleAddress:     providerData.ModuleAddress,
		})
	}

	// Transform resources
	var transformedResources []Resource
	for _, resourceData := range tofuData.RootModule.Resources {
		transformedResources = append(transformedResources, Resource{
			Address: resourceData.Address,
			Mode:    resourceData.Mode,
			Type:    resourceData.Type,
			Name:    resourceData.Name,
		})
	}

	// Transform module calls into dependencies
	var dependencies []Dependency
	for name, moduleCall := range tofuData.RootModule.ModuleCalls {
		dependencies = append(dependencies, Dependency{
			Name:              name,
			Source:            moduleCall.Source,
			VersionConstraint: moduleCall.VersionConstraint,
		})
	}

	return ModuleComponentData{
		BaseComponentData: BaseComponentData{
			Variables:   transformedVars,
			Outputs:     transformedOutputs,
			SchemaError: "",
			Readme:      false, // Will be set by caller based on actual README existence
			EditLink:    "",    // Will be set by caller
		},
		Providers:    providers,
		Dependencies: dependencies,
		Resources:    transformedResources,
	}, nil
}

func (p *Parser) buildEditLink() string {
	return fmt.Sprintf("https://github.com/%s/terraform-%s-%s/blob/%s/README.md",
		p.namespace, p.target, p.name, p.version)
}

func (p *Parser) buildRepoLink() string {
	return fmt.Sprintf("https://github.com/%s/terraform-%s-%s/tree/%s",
		p.namespace, p.target, p.name, p.version)
}

func (p *Parser) buildVCSRepository() string {
	return fmt.Sprintf("https://github.com/%s/terraform-%s-%s",
		p.namespace, p.target, p.name)
}

func (p *Parser) buildSubmoduleEditLink(submoduleName string) string {
	return fmt.Sprintf("https://github.com/%s/terraform-%s-%s/blob/%s/modules/%s/README.md",
		p.namespace, p.target, p.name, p.version, submoduleName)
}

func (p *Parser) buildExampleEditLink(exampleName string) string {
	return fmt.Sprintf("https://github.com/%s/terraform-%s-%s/blob/%s/examples/%s/README.md",
		p.namespace, p.target, p.name, p.version, exampleName)
}

func (p *Parser) buildLicenseLink(fileName string) string {
	return fmt.Sprintf("https://github.com/%s/terraform-%s-%s/blob/%s/%s",
		p.namespace, p.target, p.name, p.version, fileName)
}

func (p *Parser) hasIncompatibleLicense(licenses []license.License) bool {
	for _, lic := range licenses {
		if !lic.IsCompatible {
			return true
		}
	}
	return false
}

// formatPublishedDate returns the published date in RFC3339 format.
// Uses the actual tag creation date if available, otherwise falls back to current time.
func (p *Parser) formatPublishedDate() string {
	if p.publishedAt != nil {
		return p.publishedAt.Format(time.RFC3339)
	}
	return time.Now().Format(time.RFC3339)
}

// hasReadme checks if a README.md file exists in the given directory
func hasReadme(dir string) bool {
	readmePath := filepath.Join(dir, "README.md")
	_, err := os.Stat(readmePath)
	return err == nil
}
