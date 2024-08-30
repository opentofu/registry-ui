package providerdocsource

import (
	"context"
	"fmt"

	"github.com/opentofu/libregistry/types/provider"
	"github.com/opentofu/registry-ui/internal/providerindex/providerindexstorage"
	"github.com/opentofu/registry-ui/internal/providerindex/providertypes"
)

func newProviderDoc() *providerDoc {
	return &providerDoc{
		root:        nil,
		resources:   map[string]DocumentationItem{},
		datasources: map[string]DocumentationItem{},
		functions:   map[string]DocumentationItem{},
		guides:      map[string]DocumentationItem{},
	}
}

type providerDoc struct {
	root        DocumentationItem
	resources   map[string]DocumentationItem
	datasources map[string]DocumentationItem
	functions   map[string]DocumentationItem
	guides      map[string]DocumentationItem
}

func (p providerDoc) Store(ctx context.Context, addr provider.Addr, version providertypes.ProviderVersionDescriptor, storage providerindexstorage.API, language providertypes.CDKTFLanguage) error {
	if p.root != nil {
		if err := p.root.Store(ctx, addr, version, storage, language, providertypes.DocItemKindRoot, "index"); err != nil {
			return fmt.Errorf("failed to store root documentation (%w)", err)
		}
	}
	if err := p.storeKind(ctx, addr, version, storage, language, providertypes.DocItemKindResource, p.resources); err != nil {
		return fmt.Errorf("failed to store resources (%w)", err)
	}
	if err := p.storeKind(ctx, addr, version, storage, language, providertypes.DocItemKindDataSource, p.datasources); err != nil {
		return fmt.Errorf("failed to store resources (%w)", err)
	}
	if err := p.storeKind(ctx, addr, version, storage, language, providertypes.DocItemKindFunction, p.functions); err != nil {
		return fmt.Errorf("failed to store resources (%w)", err)
	}
	if err := p.storeKind(ctx, addr, version, storage, language, providertypes.DocItemKindGuide, p.guides); err != nil {
		return fmt.Errorf("failed to store resources (%w)", err)
	}
	return nil
}

func (p providerDoc) ToProviderTypes(ctx context.Context) providertypes.ProviderDocs {
	var root *providertypes.ProviderDocItem
	if p.root != nil && !p.root.IsError() {
		rootType := p.root.ToProviderTypes(ctx)
		root = &rootType
	}

	return providertypes.ProviderDocs{
		Root:        root,
		Resources:   p.convertDocItems(ctx, p.resources),
		DataSources: p.convertDocItems(ctx, p.datasources),
		Functions:   p.convertDocItems(ctx, p.functions),
		Guides:      p.convertDocItems(ctx, p.guides),
	}
}

func (p providerDoc) Root() (DocumentationItem, error) {
	return p.root, nil
}

func (p providerDoc) Resources() (map[string]DocumentationItem, error) {
	return p.resources, nil
}

func (p providerDoc) DataSources() (map[string]DocumentationItem, error) {
	return p.datasources, nil
}

func (p providerDoc) Functions() (map[string]DocumentationItem, error) {
	return p.functions, nil
}

func (p providerDoc) Guides() (map[string]DocumentationItem, error) {
	return p.guides, nil
}

func (p providerDoc) convertDocItems(ctx context.Context, items map[string]DocumentationItem) []providertypes.ProviderDocItem {
	//goland:noinspection GoPreferNilSlice
	result := []providertypes.ProviderDocItem{}
	for _, value := range items {
		result = append(result, value.ToProviderTypes(ctx))
	}
	return result
}

func (p providerDoc) storeKind(ctx context.Context, addr provider.Addr, version providertypes.ProviderVersionDescriptor, storage providerindexstorage.API, language providertypes.CDKTFLanguage, kind providertypes.DocItemKind, data map[string]DocumentationItem) error {
	for name, item := range data {
		if err := item.Store(ctx, addr, version, storage, language, kind, providertypes.DocItemName(name)); err != nil {
			return fmt.Errorf("failed to store %s (%w)", name, err)
		}
	}
	return nil
}
