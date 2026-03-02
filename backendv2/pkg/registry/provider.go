package registry

import (
	"context"
	"fmt"
	"log/slog"
	"path/filepath"
	"sort"
	"strings"

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

func (r *Client) ListProviders(ctx context.Context, filter string) ([]Provider, error) {
	ctx, span := telemetry.Tracer().Start(ctx, "registry.list_providers")
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
			slog.DebugContext(ctx, "Failed to list namespaces in letter directory", "path", letterPath, "error", err)
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
				slog.DebugContext(ctx, "Failed to list provider JSON files", "path", namespacePath, "error", err)
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

func (r *Client) GetProvider(ctx context.Context, namespace, name string) (*Provider, error) {
	ctx, span := telemetry.Tracer().Start(ctx, "registry.get_provider")
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
