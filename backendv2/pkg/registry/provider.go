package registry

import (
	"fmt"
	"path/filepath"
	"sort"
	"strings"
	"time"
)

type Provider struct {
	Namespace   string   `json:"namespace"`
	Name        string   `json:"name"`
	Description string   `json:"description,omitempty"`
	Link        string   `json:"link,omitempty"`
	Versions    []string `json:"versions,omitempty"`
}

type ProviderVersion struct {
	Namespace    string     `json:"namespace"`
	Name         string     `json:"name"`
	Version      string     `json:"version"`
	Protocols    []string   `json:"protocols,omitempty"`
	Platforms    []Platform `json:"platforms,omitempty"`
	SHASumsURL   string     `json:"shasums_url,omitempty"`
	SignatureURL string     `json:"shasums_signature_url,omitempty"`
	Published    time.Time  `json:"published,omitempty"`
}

type Platform struct {
	OS       string `json:"os"`
	Arch     string `json:"arch"`
	Filename string `json:"filename"`
	URL      string `json:"download_url"`
	SHA256   string `json:"shasum"`
}

type providerJSON struct {
	Versions []struct {
		Version      string   `json:"version"`
		Protocols    []string `json:"protocols"`
		SHASumsURL   string   `json:"shasums_url"`
		SignatureURL string   `json:"shasums_signature_url"`
		Targets      []struct {
			OS       string `json:"os"`
			Arch     string `json:"arch"`
			Filename string `json:"filename"`
			URL      string `json:"download_url"`
			SHA256   string `json:"shasum"`
		} `json:"targets"`
	} `json:"versions"`
}

func (r *Registry) ListProviders(filter string) ([]Provider, error) {
	providersDir := filepath.Join(r.path, "providers")
	if !fileExists(providersDir) {
		return nil, fmt.Errorf("providers directory not found")
	}

	filterParts := parseFilter(filter)
	var providers []Provider

	letterDirs, err := listDirectories(providersDir)
	if err != nil {
		return nil, fmt.Errorf("failed to list provider directories: %w", err)
	}

	for _, letter := range letterDirs {
		letterPath := filepath.Join(providersDir, letter)
		namespaces, err := listDirectories(letterPath)
		if err != nil {
			continue
		}

		for _, namespace := range namespaces {
			if len(filterParts) > 0 && !matchPattern(filterParts[0], namespace) {
				continue
			}

			namespacePath := filepath.Join(letterPath, namespace)
			providerNames, err := listJSONFiles(namespacePath)
			if err != nil {
				continue
			}

			for _, name := range providerNames {
				parts := []string{namespace, name}
				if !matchesFilter(parts, filterParts) {
					continue
				}

				provider := Provider{
					Namespace: namespace,
					Name:      name,
				}

				jsonPath := filepath.Join(namespacePath, name+".json")
				var data providerJSON
				if err := readJSONFile(jsonPath, &data); err == nil {
					for _, v := range data.Versions {
						provider.Versions = append(provider.Versions, v.Version)
					}
					provider.Link = fmt.Sprintf("https://github.com/%s/terraform-provider-%s", namespace, name)
				}

				providers = append(providers, provider)
			}
		}
	}

	sort.Slice(providers, func(i, j int) bool {
		if providers[i].Namespace != providers[j].Namespace {
			return providers[i].Namespace < providers[j].Namespace
		}
		return providers[i].Name < providers[j].Name
	})

	return providers, nil
}

func (r *Registry) ListProviderVersions(filter string) ([]ProviderVersion, error) {
	providers, err := r.ListProviders(filter)
	if err != nil {
		return nil, err
	}

	var versions []ProviderVersion
	for _, provider := range providers {
		providerVersions, err := r.getProviderVersions(provider.Namespace, provider.Name)
		if err != nil {
			continue
		}
		versions = append(versions, providerVersions...)
	}

	return versions, nil
}

func (r *Registry) GetProvider(namespace, name string) (*Provider, error) {
	firstLetter := strings.ToLower(string(namespace[0]))
	if firstLetter >= "0" && firstLetter <= "9" {
		firstLetter = string(namespace[0])
	}

	jsonPath := filepath.Join(r.path, "providers", firstLetter, namespace, name+".json")
	if !fileExists(jsonPath) {
		return nil, fmt.Errorf("provider %s/%s not found", namespace, name)
	}

	var data providerJSON
	if err := readJSONFile(jsonPath, &data); err != nil {
		return nil, fmt.Errorf("failed to read provider data: %w", err)
	}

	provider := &Provider{
		Namespace: namespace,
		Name:      name,
		Link:      fmt.Sprintf("https://github.com/%s/terraform-provider-%s", namespace, name),
	}

	for _, v := range data.Versions {
		provider.Versions = append(provider.Versions, v.Version)
	}

	return provider, nil
}

func (r *Registry) GetProviderVersion(namespace, name, version string) (*ProviderVersion, error) {
	firstLetter := strings.ToLower(string(namespace[0]))
	if firstLetter >= "0" && firstLetter <= "9" {
		firstLetter = string(namespace[0])
	}

	jsonPath := filepath.Join(r.path, "providers", firstLetter, namespace, name+".json")
	if !fileExists(jsonPath) {
		return nil, fmt.Errorf("provider %s/%s not found", namespace, name)
	}

	var data providerJSON
	if err := readJSONFile(jsonPath, &data); err != nil {
		return nil, fmt.Errorf("failed to read provider data: %w", err)
	}

	for _, v := range data.Versions {
		if v.Version == version {
			pv := &ProviderVersion{
				Namespace:    namespace,
				Name:         name,
				Version:      v.Version,
				Protocols:    v.Protocols,
				SHASumsURL:   v.SHASumsURL,
				SignatureURL: v.SignatureURL,
			}

			for _, target := range v.Targets {
				pv.Platforms = append(pv.Platforms, Platform{
					OS:       target.OS,
					Arch:     target.Arch,
					Filename: target.Filename,
					URL:      target.URL,
					SHA256:   target.SHA256,
				})
			}

			return pv, nil
		}
	}

	return nil, fmt.Errorf("version %s not found for provider %s/%s", version, namespace, name)
}

func (r *Registry) getProviderVersions(namespace, name string) ([]ProviderVersion, error) {
	firstLetter := strings.ToLower(string(namespace[0]))
	if firstLetter >= "0" && firstLetter <= "9" {
		firstLetter = string(namespace[0])
	}

	jsonPath := filepath.Join(r.path, "providers", firstLetter, namespace, name+".json")
	if !fileExists(jsonPath) {
		return nil, fmt.Errorf("provider %s/%s not found", namespace, name)
	}

	var data providerJSON
	if err := readJSONFile(jsonPath, &data); err != nil {
		return nil, fmt.Errorf("failed to read provider data: %w", err)
	}

	var versions []ProviderVersion
	for _, v := range data.Versions {
		pv := ProviderVersion{
			Namespace:    namespace,
			Name:         name,
			Version:      v.Version,
			Protocols:    v.Protocols,
			SHASumsURL:   v.SHASumsURL,
			SignatureURL: v.SignatureURL,
		}

		for _, target := range v.Targets {
			pv.Platforms = append(pv.Platforms, Platform{
				OS:       target.OS,
				Arch:     target.Arch,
				Filename: target.Filename,
				URL:      target.URL,
				SHA256:   target.SHA256,
			})
		}

		versions = append(versions, pv)
	}

	return versions, nil
}
