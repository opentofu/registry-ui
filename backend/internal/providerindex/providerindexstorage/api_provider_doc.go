package providerindexstorage

import (
	"context"
	"path"

	"github.com/opentofu/libregistry/types/provider"
	"github.com/opentofu/registry-ui/internal/indexstorage"
)

func (s storage) getProviderDocPath(_ context.Context, providerAddr provider.Addr, version provider.VersionNumber) indexstorage.Path {
	providerAddr = providerAddr.Normalize()
	version = version.Normalize()
	return indexstorage.Path(path.Join(providerAddr.Namespace, providerAddr.Name, string(version), "index.md"))
}

func (s storage) GetProviderDoc(ctx context.Context, providerAddr provider.Addr, version provider.VersionNumber) ([]byte, error) {
	// TODO validate provider addr
	if err := version.Validate(); err != nil {
		return nil, err
	}
	// TODO add typed errors
	return s.indexStorageAPI.ReadFile(ctx, s.getProviderDocPath(ctx, providerAddr, version))
}

func (s storage) StoreProviderDoc(ctx context.Context, providerAddr provider.Addr, version provider.VersionNumber, data []byte) error {
	// TODO validate provider addr
	if err := version.Validate(); err != nil {
		return err
	}
	// TODO add typed errors
	return s.indexStorageAPI.WriteFile(ctx, s.getProviderDocPath(ctx, providerAddr, version), data)
}
