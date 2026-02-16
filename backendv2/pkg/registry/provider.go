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

type Provider struct {
	Namespace   string   `json:"namespace"`
	Name        string   `json:"name"`
	Description string   `json:"description,omitempty"`
	Link        string   `json:"link,omitempty"`
	Versions    []string `json:"versions,omitempty"`
	Warnings    []string `json:"warnings,omitempty"`
}

type ProviderVersion struct {
	Namespace    string     `json:"namespace"`
	Name         string     `json:"name"`
	Version      string     `json:"version"`
	Protocols    []string   `json:"protocols,omitempty"`
	Platforms    []Platform `json:"platforms,omitempty"`
	SHASumsURL   string     `json:"shasums_url,omitempty"`
	SignatureURL string     `json:"shasums_signature_url,omitempty"`
	Published    time.Time  `json:"published"`
}

type Platform struct {
	OS       string `json:"os"`
	Arch     string `json:"arch"`
	Filename string `json:"filename"`
	URL      string `json:"download_url"`
	SHA256   string `json:"shasum"`
}

type providerJSON struct {
	Warnings []string `json:"warnings"`
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

func (p *Provider) GetVersion(version string) *ProviderVersion {
	if slices.Contains(p.Versions, version) {
		return &ProviderVersion{
			Namespace: p.Namespace,
			Name:      p.Name,
			Version:   version,
		}
	}
	return nil
}

func (r *Client) ListProviders(ctx context.Context, filter string) ([]Provider, error) {
	ctx, span := telemetry.Tracer().Start(ctx, "list-providers")
	defer span.End()

	span.SetAttributes(attribute.String("filter", filter))

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
			// Skip non-matching namespaces early to avoid reading JSON files
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
				if err := readJSONFile(ctx, jsonPath, &data); err == nil {
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

	providerNames := make([]string, 0, len(providers))
	for _, provider := range providers {
		providerNames = append(providerNames, fmt.Sprintf("%s/%s", provider.Namespace, provider.Name))
	}

	span.SetAttributes(attribute.StringSlice("providers", providerNames))

	return providers, nil
}

func (r *Client) ListProviderVersions(ctx context.Context, filter string) ([]ProviderVersion, error) {
	ctx, span := telemetry.Tracer().Start(ctx, "list-provider-versions")
	defer span.End()

	span.SetAttributes(attribute.String("filter", filter))

	providers, err := r.ListProviders(ctx, filter)
	if err != nil {
		return nil, err
	}

	var versions []ProviderVersion
	for _, provider := range providers {
		providerVersions, err := r.getProviderVersions(ctx, provider.Namespace, provider.Name)
		if err != nil {
			continue
		}
		versions = append(versions, providerVersions...)
	}

	span.SetAttributes(attribute.Int("providers", len(providers)))

	return versions, nil
}

func (r *Client) GetProvider(ctx context.Context, namespace, name string) (*Provider, error) {
	ctx, span := telemetry.Tracer().Start(ctx, "get-provider")
	defer span.End()

	span.SetAttributes(attribute.String("namespace", namespace), attribute.String("name", name))

	firstLetter := strings.ToLower(string(namespace[0]))

	jsonPath := filepath.Join(r.path, "providers", firstLetter, namespace, name+".json")
	if !fileExists(jsonPath) {
		span.RecordError(fmt.Errorf("provider file not found: %s", jsonPath))
		return nil, fmt.Errorf("provider %s/%s not found", namespace, name)
	}

	var data providerJSON
	if err := readJSONFile(ctx, jsonPath, &data); err != nil {
		span.RecordError(err)
		return nil, fmt.Errorf("failed to read provider data: %w", err)
	}

	provider := &Provider{
		Namespace: namespace,
		Name:      name,
		Link:      fmt.Sprintf("https://github.com/%s/terraform-provider-%s", namespace, name),
		Warnings:  data.Warnings,
	}

	for _, v := range data.Versions {
		provider.Versions = append(provider.Versions, v.Version)
	}

	return provider, nil
}

func (r *Client) GetProviderVersion(ctx context.Context, namespace, name, version string) (*ProviderVersion, error) {
	ctx, span := telemetry.Tracer().Start(ctx, "get-provider-version")
	defer span.End()

	span.SetAttributes(attribute.String("namespace", namespace), attribute.String("name", name), attribute.String("version", version))

	firstLetter := strings.ToLower(string(namespace[0]))

	jsonPath := filepath.Join(r.path, "providers", firstLetter, namespace, name+".json")
	if !fileExists(jsonPath) {
		return nil, fmt.Errorf("provider %s/%s not found", namespace, name)
	}

	var data providerJSON
	if err := readJSONFile(ctx, jsonPath, &data); err != nil {
		return nil, fmt.Errorf("failed to read provider data: %w", err)
	}

	for _, v := range data.Versions {
		if v.Version == version {
			span.SetAttributes(attribute.Bool("found", true))

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

	span.SetAttributes(attribute.Bool("found", false))

	return nil, fmt.Errorf("version %s not found for provider %s/%s", version, namespace, name)
}

func (r *Client) getProviderVersions(ctx context.Context, namespace, name string) ([]ProviderVersion, error) {
	firstLetter := strings.ToLower(string(namespace[0]))

	jsonPath := filepath.Join(r.path, "providers", firstLetter, namespace, name+".json")
	if !fileExists(jsonPath) {
		return nil, fmt.Errorf("provider %s/%s not found", namespace, name)
	}

	var data providerJSON
	if err := readJSONFile(ctx, jsonPath, &data); err != nil {
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
