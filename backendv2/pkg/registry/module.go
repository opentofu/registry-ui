package registry

import (
	"context"
	"fmt"
	"path/filepath"
	"slices"
	"sort"
	"strings"
	"time"

	"go.opentelemetry.io/otel/attribute"

	"github.com/opentofu/registry-ui/pkg/telemetry"
)

type Module struct {
	Namespace   string   `json:"namespace"`
	Name        string   `json:"name"`
	Target      string   `json:"target"`
	Description string   `json:"description,omitempty"`
	Source      string   `json:"source,omitempty"`
	Versions    []string `json:"versions,omitempty"`
}

type ModuleVersion struct {
	Namespace string    `json:"namespace"`
	Name      string    `json:"name"`
	Target    string    `json:"target"`
	Version   string    `json:"version"`
	Source    string    `json:"source,omitempty"`
	Published time.Time `json:"published"`
}

type moduleJSON struct {
	Versions []struct {
		Version string `json:"version"`
	} `json:"versions"`
}

func (m *Module) GetVersion(version string) *ModuleVersion {
	if slices.Contains(m.Versions, version) {
		return &ModuleVersion{
			Namespace: m.Namespace,
			Name:      m.Name,
			Target:    m.Target,
			Version:   version,
			Source:    m.Source,
		}
	}
	return nil
}

func (r *Client) ListModules(ctx context.Context, filter string) ([]Module, error) {
	ctx, span := telemetry.Tracer().Start(ctx, "list-modules")
	defer span.End()

	span.SetAttributes(attribute.String("filter", filter))

	modulesDir := filepath.Join(r.path, "modules")
	if !fileExists(modulesDir) {
		return nil, fmt.Errorf("modules directory not found")
	}

	filterParts := parseFilter(filter)
	var modules []Module

	letterDirs, err := listDirectories(modulesDir)
	if err != nil {
		return nil, fmt.Errorf("failed to list module directories: %w", err)
	}

	for _, letter := range letterDirs {
		letterPath := filepath.Join(modulesDir, letter)
		namespaces, err := listDirectories(letterPath)
		if err != nil {
			continue
		}

		for _, namespace := range namespaces {
			if len(filterParts) > 0 && !matchPattern(filterParts[0], namespace) {
				continue
			}

			namespacePath := filepath.Join(letterPath, namespace)

			// List module directories (not JSON files)
			moduleDirs, err := listDirectories(namespacePath)
			if err != nil {
				continue
			}

			for _, moduleName := range moduleDirs {
				modulePath := filepath.Join(namespacePath, moduleName)

				// List JSON files in the module directory to find targets
				targetFiles, err := listJSONFiles(modulePath)
				if err != nil {
					continue
				}

				for _, targetFile := range targetFiles {
					// Target is the JSON filename without extension (e.g., aws.json -> aws)
					target := targetFile

					module := &Module{
						Namespace: namespace,
						Name:      moduleName,
						Target:    target,
					}

					parts := []string{namespace, module.Name, module.Target}

					if !matchesModuleFilter(parts, filterParts) {
						continue
					}

					jsonPath := filepath.Join(modulePath, targetFile+".json")
					var data moduleJSON
					if err := readJSONFile(ctx, jsonPath, &data); err == nil {
						for _, v := range data.Versions {
							module.Versions = append(module.Versions, v.Version)
						}
					}

					module.Source = fmt.Sprintf("https://github.com/%s/terraform-%s-%s",
						namespace, target, moduleName)

					modules = append(modules, *module)
				}
			}
		}
	}

	// use the slices package to pul out the list of module names
	sort.Slice(modules, func(i, j int) bool {
		if modules[i].Namespace != modules[j].Namespace {
			return modules[i].Namespace < modules[j].Namespace
		}
		if modules[i].Name != modules[j].Name {
			return modules[i].Name < modules[j].Name
		}
		return modules[i].Target < modules[j].Target
	})

	modulesToTrace := make([]string, len(modules))
	for _, module := range modules {
		modulesToTrace = append(modulesToTrace, fmt.Sprintf("%s/%s/%s", module.Namespace, module.Name, module.Target))
	}

	span.SetAttributes(attribute.Int("count", len(modules)), attribute.StringSlice("modules", modulesToTrace))

	return modules, nil
}

func (r *Client) ListModuleVersions(ctx context.Context, filter string) ([]ModuleVersion, error) {
	ctx, span := telemetry.Tracer().Start(ctx, "list-modules")
	defer span.End()

	span.SetAttributes(attribute.String("filter", filter))

	modules, err := r.ListModules(ctx, filter)
	if err != nil {
		return nil, err
	}

	var versions []ModuleVersion
	for _, module := range modules {
		moduleVersions, err := r.getModuleVersions(ctx, module.Namespace, module.Name, module.Target)
		if err != nil {
			continue
		}
		versions = append(versions, moduleVersions...)
	}

	span.SetAttributes()

	return versions, nil
}

func (r *Client) GetModule(ctx context.Context, namespace, name, target string) (*Module, error) {
	ctx, span := telemetry.Tracer().Start(ctx, "get-module")
	defer span.End()

	span.SetAttributes(attribute.String("namespace", namespace), attribute.String("name", name), attribute.String("target", target))

	firstLetter := strings.ToLower(string(namespace[0]))
	if firstLetter >= "0" && firstLetter <= "9" {
		firstLetter = string(namespace[0])
	}

	filename := name
	if target != "" {
		filename = fmt.Sprintf("%s/%s", name, target)
	}

	jsonPath := filepath.Join(r.path, "modules", firstLetter, namespace, filename+".json")

	if !fileExists(jsonPath) {
		filename = fmt.Sprintf("%s.%s", target, name)
		jsonPath = filepath.Join(r.path, "modules", firstLetter, namespace, filename+".json")
		if !fileExists(jsonPath) {
			span.RecordError(fmt.Errorf("module file not found: %s", jsonPath))
			return nil, fmt.Errorf("module %s/%s/%s not found", namespace, name, target)
		}
	}

	var data moduleJSON
	if err := readJSONFile(ctx, jsonPath, &data); err != nil {
		span.RecordError(err)
		return nil, fmt.Errorf("failed to read module data: %w", err)
	}

	module := &Module{
		Namespace: namespace,
		Name:      name,
		Target:    target,
		Source:    fmt.Sprintf("https://github.com/%s/terraform-%s-%s", namespace, target, name),
	}

	for _, v := range data.Versions {
		module.Versions = append(module.Versions, v.Version)
	}

	return module, nil
}

func (r *Client) GetModuleVersion(ctx context.Context, namespace, name, target, version string) (*ModuleVersion, error) {
	ctx, span := telemetry.Tracer().Start(ctx, "get-module-version")
	defer span.End()

	span.SetAttributes(attribute.String("namespace", namespace), attribute.String("name", name), attribute.String("target", target), attribute.String("version", version))

	module, err := r.GetModule(ctx, namespace, name, target)
	if err != nil {
		return nil, err
	}

	if slices.Contains(module.Versions, version) {
		return &ModuleVersion{
			Namespace: namespace,
			Name:      name,
			Target:    target,
			Version:   version,
			Source:    module.Source,
		}, nil
	}

	return nil, fmt.Errorf("version %s not found for module %s/%s/%s", version, namespace, name, target)
}

func (r *Client) getModuleVersions(ctx context.Context, namespace, name, target string) ([]ModuleVersion, error) {
	module, err := r.GetModule(ctx, namespace, name, target)
	if err != nil {
		return nil, err
	}

	var versions []ModuleVersion
	for _, v := range module.Versions {
		versions = append(versions, ModuleVersion{
			Namespace: namespace,
			Name:      name,
			Target:    target,
			Version:   v,
			Source:    module.Source,
		})
	}

	return versions, nil
}

func matchesModuleFilter(parts []string, filterParts []string) bool {
	if len(filterParts) == 0 {
		return true
	}

	if len(filterParts) == 1 {
		return matchPattern(filterParts[0], parts[0])
	}

	if len(filterParts) == 2 {
		if !matchPattern(filterParts[0], parts[0]) {
			return false
		}

		if len(parts) >= 2 && matchPattern(filterParts[1], parts[1]) {
			return true
		}
		if len(parts) >= 3 && matchPattern(filterParts[1], parts[2]) {
			return true
		}
		return false
	}

	if len(filterParts) == 3 {
		if len(parts) < 3 {
			return false
		}
		return matchPattern(filterParts[0], parts[0]) &&
			matchPattern(filterParts[1], parts[1]) &&
			matchPattern(filterParts[2], parts[2])
	}

	return false
}
