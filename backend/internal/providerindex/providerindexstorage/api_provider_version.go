package providerindexstorage

import (
	"context"
	"encoding/json"
	"os"
	"path"
	"time"

	"github.com/opentofu/libregistry/types/provider"
	"github.com/opentofu/registry-ui/internal/indexstorage"
	"github.com/opentofu/registry-ui/internal/providerindex/providertypes"
)

func (s storage) getProviderVersionPath(providerAddr provider.Addr, version provider.VersionNumber) indexstorage.Path {
	providerAddr = providerAddr.Normalize()
	version = version.Normalize()
	return indexstorage.Path(path.Join(providerAddr.Namespace, providerAddr.Name, string(version), providerAddr.Name))
}

func (s storage) getProviderVersionFile(providerAddr provider.Addr, version provider.VersionNumber) indexstorage.Path {
	providerAddr = providerAddr.Normalize()
	version = version.Normalize()
	return indexstorage.Path(path.Join(providerAddr.Namespace, providerAddr.Name, string(version), "index.json"))
}

func (s storage) GetProviderVersion(ctx context.Context, providerAddr provider.Addr, version provider.VersionNumber) (providertypes.ProviderVersion, error) {
	result := providertypes.ProviderVersion{
		ProviderVersionDescriptor: providertypes.ProviderVersionDescriptor{
			ID:        version,
			Published: time.Time{},
		},
		Docs:      providertypes.ProviderDocs{},
		CDKTFDocs: map[providertypes.CDKTFLanguage]providertypes.ProviderDocs{},
		Licenses:  nil,
	}

	// TODO validate provider addr
	if err := version.Validate(); err != nil {
		return result, err
	}

	indexContents, err := s.indexStorageAPI.ReadFile(ctx, s.getProviderVersionFile(providerAddr, version))
	if err != nil {
		if os.IsNotExist(err) {
			return result, &ProviderVersionNotFoundError{BaseError: BaseError{Cause: err}, ProviderAddr: providerAddr, Version: version}
		}
		return result, &ProviderVersionReadFailedError{BaseError: BaseError{Cause: err}, ProviderAddr: providerAddr, Version: version}
	}

	if err := json.Unmarshal(indexContents, &result); err != nil {
		return result, &ProviderVersionReadFailedError{BaseError: BaseError{Cause: err}, ProviderAddr: providerAddr, Version: version}
	}
	return result, nil
}

func (s storage) StoreProviderVersion(ctx context.Context, providerAddr provider.Addr, providerVersion providertypes.ProviderVersion) error {
	// TODO validate provider addr
	if err := providerVersion.ID.Validate(); err != nil {
		return err
	}
	marshalled, err := json.Marshal(providerVersion)
	if err != nil {
		return &ProviderVersionStoreFailedError{BaseError: BaseError{Cause: err}, ProviderAddr: providerAddr, Version: providerVersion.ID}
	}
	if err := s.indexStorageAPI.WriteFile(ctx, s.getProviderVersionFile(providerAddr, providerVersion.ID), marshalled); err != nil {
		return &ProviderVersionStoreFailedError{BaseError: BaseError{Cause: err}, ProviderAddr: providerAddr, Version: providerVersion.ID}
	}
	return nil
}

func (s storage) DeleteProviderVersion(ctx context.Context, providerAddr provider.Addr, version provider.VersionNumber) error {
	// TODO validate provider addr
	if err := version.Validate(); err != nil {
		return err
	}
	return s.indexStorageAPI.RemoveAll(ctx, s.getProviderVersionPath(providerAddr, version))
}
