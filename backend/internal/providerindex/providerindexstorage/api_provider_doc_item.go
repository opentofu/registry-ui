package providerindexstorage

import (
	"context"
	"path"

	"github.com/opentofu/registry-ui/internal/indexstorage"
	"github.com/opentofu/registry-ui/internal/providerindex/providertypes"
)

func (s storage) getProviderDocItemPath(_ context.Context, providerAddr providertypes.ProviderAddr, version string, kind providertypes.DocItemKind, name providertypes.DocItemName) indexstorage.Path {
	name = name.Normalize()
	if kind == providertypes.DocItemKindRoot {
		return indexstorage.Path(path.Join(providerAddr.Namespace, providerAddr.Name, string(version), string(name)+".md"))
	}
	return indexstorage.Path(path.Join(providerAddr.Namespace, providerAddr.Name, string(version), string(kind)+"s", string(name)+".md"))
}

func (s storage) GetProviderDocItem(ctx context.Context, providerAddr providertypes.ProviderAddr, version string, kind providertypes.DocItemKind, name providertypes.DocItemName) ([]byte, error) {
	if err := kind.Validate(); err != nil {
		return nil, err
	}
	if err := name.Validate(); err != nil {
		return nil, err
	}
	return s.indexStorageAPI.ReadFile(ctx, s.getProviderDocItemPath(ctx, providerAddr, version, kind, name))
}

func (s storage) StoreProviderDocItem(ctx context.Context, providerAddr providertypes.ProviderAddr, version string, kind providertypes.DocItemKind, name providertypes.DocItemName, data []byte) error {
	if err := kind.Validate(); err != nil {
		return err
	}
	if err := name.Validate(); err != nil {
		return err
	}
	return s.indexStorageAPI.WriteFile(ctx, s.getProviderDocItemPath(ctx, providerAddr, version, kind, name), data)
}
