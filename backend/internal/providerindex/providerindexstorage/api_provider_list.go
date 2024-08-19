package providerindexstorage

import (
	"context"
	"encoding/json"
	"os"

	"github.com/opentofu/registry-ui/internal/indexstorage"
	"github.com/opentofu/registry-ui/internal/providerindex/providertypes"
)

func (s storage) getProviderListFile() indexstorage.Path {
	return "index.json"
}

func (s storage) GetProviderList(ctx context.Context) (providertypes.ProviderList, error) {
	index := providertypes.ProviderList{
		Providers: []*providertypes.Provider{},
	}
	indexContents, err := s.indexStorageAPI.ReadFile(ctx, s.getProviderListFile())
	if err != nil {
		if os.IsNotExist(err) {
			return index, &ProviderListNotFoundError{BaseError: BaseError{Cause: err}}
		}
		return index, &ProviderListReadFailedError{BaseError: BaseError{Cause: err}}
	}

	if err := json.Unmarshal(indexContents, &index); err != nil {
		return index, &ProviderListReadFailedError{BaseError: BaseError{Cause: err}}
	}
	return index, nil
}

func (s storage) StoreProviderList(ctx context.Context, providerList providertypes.ProviderList) error {
	marshalled, err := json.Marshal(providerList)
	if err != nil {
		return &ProviderListStoreFailedError{BaseError: BaseError{Cause: err}}
	}
	if err := s.indexStorageAPI.WriteFile(ctx, s.getProviderListFile(), marshalled); err != nil {
		return &ProviderListStoreFailedError{BaseError: BaseError{Cause: err}}
	}
	return nil
}
