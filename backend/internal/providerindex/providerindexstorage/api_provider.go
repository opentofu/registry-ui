package providerindexstorage

import (
	"context"
	"encoding/json"
	"os"
	"path"

	"github.com/opentofu/libregistry/types/provider"
	"github.com/opentofu/registry-ui/internal/indexstorage"
	"github.com/opentofu/registry-ui/internal/providerindex/providertypes"
)

func (s storage) getProviderPath(providerAddr provider.Addr) indexstorage.Path {
	providerAddr = providerAddr.Normalize()
	return indexstorage.Path(path.Join(providerAddr.Namespace, providerAddr.Name))
}

func (s storage) getProviderFile(providerAddr provider.Addr) indexstorage.Path {
	providerAddr = providerAddr.Normalize()
	return indexstorage.Path(path.Join(providerAddr.Namespace, providerAddr.Name, "index.json"))
}

func (s storage) GetProvider(ctx context.Context, providerAddr provider.Addr) (providertypes.Provider, error) {
	// TODO validate provider addr
	index := providertypes.Provider{
		Addr:        providertypes.Addr(providerAddr),
		Description: "",
		Versions:    []providertypes.ProviderVersionDescriptor{},
	}
	indexContents, err := s.indexStorageAPI.ReadFile(ctx, s.getProviderFile(providerAddr))
	if err != nil {
		if os.IsNotExist(err) {
			return index, &ProviderNotFoundError{BaseError: BaseError{Cause: err}, ProviderAddr: providerAddr}
		}
		return index, &ProviderReadFailedError{BaseError: BaseError{Cause: err}, ProviderAddr: providerAddr}
	}

	if err := json.Unmarshal(indexContents, &index); err != nil {
		return index, &ProviderReadFailedError{BaseError: BaseError{Cause: err}, ProviderAddr: providerAddr}
	}
	return index, nil
}

func (s storage) StoreProvider(ctx context.Context, provider providertypes.Provider) error {
	// TODO validate provider addr
	marshalled, err := json.Marshal(provider)
	if err != nil {
		return &ProviderStoreFailedError{BaseError: BaseError{Cause: err}, ProviderAddr: provider.Addr.Addr}
	}
	if err := s.indexStorageAPI.WriteFile(ctx, s.getProviderFile(provider.Addr.Addr), marshalled); err != nil {
		return &ProviderStoreFailedError{BaseError: BaseError{Cause: err}, ProviderAddr: provider.Addr.Addr}
	}
	return nil
}

func (s storage) DeleteProvider(ctx context.Context, providerAddr provider.Addr) error {
	// TODO validate provider addr
	return s.indexStorageAPI.RemoveAll(ctx, s.getProviderPath(providerAddr))
}
