package providerdocsource

import (
	"context"

	"github.com/opentofu/libregistry/types/provider"
	"github.com/opentofu/libregistry/vcs"
	"github.com/opentofu/registry-ui/internal/license"
	"github.com/opentofu/registry-ui/internal/providerindex/providerindexstorage"
	"github.com/opentofu/registry-ui/internal/providerindex/providertypes"
)

type API interface {
	Describe(ctx context.Context, workingCopy vcs.WorkingCopy) (ProviderDocumentation, error)
}

type ProviderDocumentation interface {
	GetDocumentation(ctx context.Context) (Documentation, error)
	GetCDKTF(ctx context.Context) (map[string]Documentation, error)
	GetLicenses(ctx context.Context) (license.List, error)

	// Store writes the provider documentation to the storage.
	Store(ctx context.Context, addr provider.Addr, version providertypes.ProviderVersionDescriptor, storage providerindexstorage.API) (providertypes.ProviderVersion, error)

	ToProviderTypes(ctx context.Context, version providertypes.ProviderVersionDescriptor) providertypes.ProviderVersion
}

type Documentation interface {
	// Root will return the root DocumentationItem, if any. This function may return nil without an error.
	Root() (DocumentationItem, error)
	Resources() (map[string]DocumentationItem, error)
	DataSources() (map[string]DocumentationItem, error)
	Functions() (map[string]DocumentationItem, error)

	// Store stores the documentation. The language may have the special value of empty, indicating a non-CDKTF doc.
	Store(ctx context.Context, addr provider.Addr, version providertypes.ProviderVersionDescriptor, storage providerindexstorage.API, language providertypes.CDKTFLanguage) error

	ToProviderTypes(ctx context.Context) providertypes.ProviderDocs
}

type DocumentationItem interface {
	GetName(ctx context.Context) (string, error)

	GetTitle(ctx context.Context) (string, error)
	GetSubcategory(ctx context.Context) (string, error)
	GetDescription(ctx context.Context) (string, error)

	GetContents(ctx context.Context) ([]byte, error)

	// Store stores the documentation item. The language may have the special value of empty, indicating a non-CDKTF doc item.
	Store(ctx context.Context, addr provider.Addr, version providertypes.ProviderVersionDescriptor, storage providerindexstorage.API, language providertypes.CDKTFLanguage, itemKind providertypes.DocItemKind, itemName providertypes.DocItemName) error

	ToProviderTypes(ctx context.Context) providertypes.ProviderDocItem
}
