package registry

import (
	"fmt"
	"path/filepath"
	"sort"
	"strings"
	"time"
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
	Published time.Time `json:"published,omitempty"`
}

type moduleJSON struct {
	Versions []struct {
		Version string `json:"version"`
	} `json:"versions"`
}

func (r *Registry) ListModules(filter string) ([]Module, error) {
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
			moduleFiles, err := listJSONFiles(namespacePath)
			if err != nil {
				continue
			}

			for _, moduleFile := range moduleFiles {
				module, err := r.parseModuleFromFilename(namespace, moduleFile)
				if err != nil {
					continue
				}

				parts := []string{namespace, module.Name}
				if module.Target != "" {
					parts = append(parts, module.Target)
				}

				if !matchesModuleFilter(parts, filterParts) {
					continue
				}

				jsonPath := filepath.Join(namespacePath, moduleFile+".json")
				var data moduleJSON
				if err := readJSONFile(jsonPath, &data); err == nil {
					for _, v := range data.Versions {
						module.Versions = append(module.Versions, v.Version)
					}
				}

				module.Source = fmt.Sprintf("https://github.com/%s/terraform-%s-module-%s",
					namespace, module.Target, module.Name)

				modules = append(modules, *module)
			}
		}
	}

	sort.Slice(modules, func(i, j int) bool {
		if modules[i].Namespace != modules[j].Namespace {
			return modules[i].Namespace < modules[j].Namespace
		}
		if modules[i].Name != modules[j].Name {
			return modules[i].Name < modules[j].Name
		}
		return modules[i].Target < modules[j].Target
	})

	return modules, nil
}

func (r *Registry) ListModuleVersions(filter string) ([]ModuleVersion, error) {
	modules, err := r.ListModules(filter)
	if err != nil {
		return nil, err
	}

	var versions []ModuleVersion
	for _, module := range modules {
		moduleVersions, err := r.getModuleVersions(module.Namespace, module.Name, module.Target)
		if err != nil {
			continue
		}
		versions = append(versions, moduleVersions...)
	}

	return versions, nil
}

func (r *Registry) GetModule(namespace, name, target string) (*Module, error) {
	firstLetter := strings.ToLower(string(namespace[0]))
	if firstLetter >= "0" && firstLetter <= "9" {
		firstLetter = string(namespace[0])
	}

	filename := name
	if target != "" {
		filename = fmt.Sprintf("%s.%s", name, target)
	}

	jsonPath := filepath.Join(r.path, "modules", firstLetter, namespace, filename+".json")
	if !fileExists(jsonPath) {
		filename = fmt.Sprintf("%s.%s", target, name)
		jsonPath = filepath.Join(r.path, "modules", firstLetter, namespace, filename+".json")
		if !fileExists(jsonPath) {
			return nil, fmt.Errorf("module %s/%s/%s not found", namespace, name, target)
		}
	}

	var data moduleJSON
	if err := readJSONFile(jsonPath, &data); err != nil {
		return nil, fmt.Errorf("failed to read module data: %w", err)
	}

	module := &Module{
		Namespace: namespace,
		Name:      name,
		Target:    target,
		Source:    fmt.Sprintf("https://github.com/%s/terraform-%s-module-%s", namespace, target, name),
	}

	for _, v := range data.Versions {
		module.Versions = append(module.Versions, v.Version)
	}

	return module, nil
}

func (r *Registry) GetModuleVersion(namespace, name, target, version string) (*ModuleVersion, error) {
	module, err := r.GetModule(namespace, name, target)
	if err != nil {
		return nil, err
	}

	for _, v := range module.Versions {
		if v == version {
			return &ModuleVersion{
				Namespace: namespace,
				Name:      name,
				Target:    target,
				Version:   version,
				Source:    module.Source,
			}, nil
		}
	}

	return nil, fmt.Errorf("version %s not found for module %s/%s/%s", version, namespace, name, target)
}

func (r *Registry) getModuleVersions(namespace, name, target string) ([]ModuleVersion, error) {
	module, err := r.GetModule(namespace, name, target)
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

func (r *Registry) parseModuleFromFilename(namespace, filename string) (*Module, error) {
	parts := strings.Split(filename, ".")

	module := &Module{
		Namespace: namespace,
	}

	if len(parts) == 1 {
		module.Name = parts[0]
	} else if len(parts) == 2 {
		module.Name = parts[0]
		module.Target = parts[1]
	} else {
		module.Name = strings.Join(parts[:len(parts)-1], ".")
		module.Target = parts[len(parts)-1]
	}

	return module, nil
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
