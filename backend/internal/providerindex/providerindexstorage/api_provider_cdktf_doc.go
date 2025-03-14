package providerindexstorage

import (
	"context"
	"path"

	"github.com/opentofu/registry-ui/internal/indexstorage"
	"github.com/opentofu/registry-ui/internal/providerindex/providertypes"
)

func (s storage) getProviderCDKTFDocPath(_ context.Context, providerAddr providertypes.ProviderAddr, version string, language providertypes.CDKTFLanguage) indexstorage.Path {
	return indexstorage.Path(path.Join(providerAddr.Namespace, providerAddr.Name, string(version), cdktfDirName, string(language), "index.md"))
}

func (s storage) GetProviderCDKTFDoc(ctx context.Context, providerAddr providertypes.ProviderAddr, version string, language providertypes.CDKTFLanguage) ([]byte, error) {
	if err := language.Validate(); err != nil {
		return nil, err
	}
	// TODO add typed errors
	return s.indexStorageAPI.ReadFile(ctx, s.getProviderCDKTFDocPath(ctx, providerAddr, version, language))
}

func (s storage) StoreProviderCDKTFDoc(ctx context.Context, providerAddr providertypes.ProviderAddr, version string, language providertypes.CDKTFLanguage, data []byte) error {
	if err := language.Validate(); err != nil {
		return err
	}
	// TODO add typed errors
	return s.indexStorageAPI.WriteFile(ctx, s.getProviderCDKTFDocPath(ctx, providerAddr, version, language), data)
}
