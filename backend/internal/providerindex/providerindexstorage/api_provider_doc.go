package providerindexstorage

import (
	"context"
	"path"

	"github.com/opentofu/registry-ui/internal/indexstorage"
	"github.com/opentofu/registry-ui/internal/providerindex/providertypes"
)

func (s storage) getProviderDocPath(_ context.Context, providerAddr providertypes.ProviderAddr, version string) indexstorage.Path {
	return indexstorage.Path(path.Join(providerAddr.Namespace, providerAddr.Name, string(version), "index.md"))
}

func (s storage) GetProviderDoc(ctx context.Context, providerAddr providertypes.ProviderAddr, version string) ([]byte, error) {
	// TODO add typed errors
	return s.indexStorageAPI.ReadFile(ctx, s.getProviderDocPath(ctx, providerAddr, version))
}

func (s storage) StoreProviderDoc(ctx context.Context, providerAddr providertypes.ProviderAddr, version string, data []byte) error {
	// TODO add typed errors
	return s.indexStorageAPI.WriteFile(ctx, s.getProviderDocPath(ctx, providerAddr, version), data)
}
