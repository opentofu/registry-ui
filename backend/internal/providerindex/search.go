package providerindex

import (
	"context"
	"fmt"

	"github.com/opentofu/libregistry/types/module"
	"github.com/opentofu/libregistry/types/provider"
	"github.com/opentofu/registry-ui/internal/providerindex/providertypes"
	"github.com/opentofu/registry-ui/internal/search"
	"github.com/opentofu/registry-ui/internal/search/searchtypes"
)

const indexPrefix = "providers"

type providerSearch struct {
	searchAPI search.API
}

func (p providerSearch) indexProviderVersion(ctx context.Context, providerAddr provider.Addr, providerDetails *providertypes.Provider, providerVersionDetails providertypes.ProviderVersion) error {
	version := providerVersionDetails.ProviderVersionDescriptor.ID
	popularity := providerDetails.Popularity
	if providerAddr.ToRepositoryAddr() == providerDetails.ForkOf.ToRepositoryAddr() {
		// If the non-canonical repo address matches where we forked from, we take the popularity of the upstream.
		popularity = providerDetails.UpstreamPopularity
	}
	providerItem := searchtypes.IndexItem{
		ID:          searchtypes.IndexID("providers/" + providerAddr.String()),
		Type:        searchtypes.IndexTypeProvider,
		Addr:        providerAddr.String(),
		Version:     string(version),
		Title:       providerAddr.Name,
		Description: "",
		LinkVariables: map[string]string{
			"namespace": providerAddr.Namespace,
			"name":      providerAddr.Name,
			"version":   string(version),
		},
		ParentID:   "",
		Popularity: popularity,
		Warnings:   len(providerDetails.Warnings),
	}

	if err := p.searchAPI.AddItem(ctx, providerItem); err != nil {
		return fmt.Errorf("failed to add provider %s search index (%w)", providerAddr, err)
	}

	for _, item := range []struct {
		typeName  string
		indexType searchtypes.IndexType
		items     []providertypes.ProviderDocItem
	}{
		{
			"resource",
			searchtypes.IndexTypeProviderResource,
			providerVersionDetails.Docs.Resources,
		},
		{
			"datasource",
			searchtypes.IndexTypeProviderDatasource,
			providerVersionDetails.Docs.DataSources,
		},
		{
			"function",
			searchtypes.IndexTypeProviderFunction,
			providerVersionDetails.Docs.Functions,
		},
	} {
		for _, docItem := range item.items {
			title := docItem.Title
			if err := p.searchAPI.AddItem(ctx, searchtypes.IndexItem{
				ID:          searchtypes.IndexID("providers/" + providerAddr.String() + "/" + item.typeName + "s/" + string(docItem.Name)),
				Type:        item.indexType,
				Addr:        providerAddr.String(),
				Version:     string(version),
				Title:       title,
				Description: docItem.Description,
				LinkVariables: map[string]string{
					"namespace": providerAddr.Namespace,
					"name":      providerAddr.Name,
					"version":   string(version),
					"id":        string(docItem.Name),
				},
				ParentID:   providerItem.ID,
				Popularity: popularity,
				Warnings:   len(providerDetails.Warnings),
			}); err != nil {
				return fmt.Errorf("failed to add resource %s to search index (%w)", docItem.Name, err)
			}
		}
	}
	return nil
}

func (p providerSearch) removeProviderVersionFromSearchIndex(ctx context.Context, addr provider.Addr, version provider.VersionNumber) error {
	for _, t := range []searchtypes.IndexType{
		searchtypes.IndexTypeProvider,
		searchtypes.IndexTypeProviderResource,
		searchtypes.IndexTypeProviderDatasource,
	} {
		if err := p.searchAPI.RemoveVersionItems(ctx, t, addr.String(), string(version)); err != nil {
			return fmt.Errorf("failed to remove provider %s version %s (%w)", addr, version, err)
		}
	}
	return nil
}

func (p providerSearch) removeModuleFromSearchIndex(ctx context.Context, addr module.Addr) error { //nolint:unused
	return p.searchAPI.RemoveItem(ctx, searchtypes.IndexID(indexPrefix+"/"+addr.String()))
}
