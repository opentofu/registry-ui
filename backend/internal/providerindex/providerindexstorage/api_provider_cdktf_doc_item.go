package providerindexstorage

import (
	"context"
	"path"

	"github.com/opentofu/libregistry/types/provider"
	"github.com/opentofu/registry-ui/internal/indexstorage"
	"github.com/opentofu/registry-ui/internal/providerindex/providertypes"
)

func (s storage) getProviderCDKTFDocItemPath(_ context.Context, providerAddr provider.Addr, version provider.VersionNumber, language providertypes.CDKTFLanguage, kind providertypes.DocItemKind, name providertypes.DocItemName) indexstorage.Path {
	providerAddr = providerAddr.Normalize()
	version = version.Normalize()
	name = name.Normalize()
	return indexstorage.Path(path.Join(providerAddr.Namespace, providerAddr.Name, string(version), cdktfDirName, string(language), string(kind)+"s", string(name)+".md"))
}

func (s storage) GetProviderCDKTFDocItem(ctx context.Context, providerAddr provider.Addr, version provider.VersionNumber, language providertypes.CDKTFLanguage, kind providertypes.DocItemKind, name providertypes.DocItemName) ([]byte, error) {
	// TODO validate provider addr
	if err := version.Validate(); err != nil {
		return nil, err
	}
	if err := kind.Validate(); err != nil {
		return nil, err
	}
	if err := language.Validate(); err != nil {
		return nil, err
	}
	if err := name.Validate(); err != nil {
		return nil, err
	}
	return s.indexStorageAPI.ReadFile(ctx, s.getProviderCDKTFDocItemPath(ctx, providerAddr, version, language, kind, name))
}

func (s storage) StoreProviderCDKTFDocItem(ctx context.Context, providerAddr provider.Addr, version provider.VersionNumber, language providertypes.CDKTFLanguage, kind providertypes.DocItemKind, name providertypes.DocItemName, data []byte) error {
	if err := version.Validate(); err != nil {
		return err
	}
	if err := kind.Validate(); err != nil {
		return err
	}
	if err := language.Validate(); err != nil {
		return err
	}
	if err := name.Validate(); err != nil {
		return err
	}
	return s.indexStorageAPI.WriteFile(ctx, s.getProviderCDKTFDocItemPath(ctx, providerAddr, version, language, kind, name), data)
}
