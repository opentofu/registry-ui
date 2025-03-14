package providerindexstorage

import (
	"context"
	"path"

	"github.com/opentofu/registry-ui/internal/indexstorage"
	"github.com/opentofu/registry-ui/internal/providerindex/providertypes"
)

func (s storage) getProviderCDKTFDocItemPath(_ context.Context, providerAddr providertypes.ProviderAddr, version string, language providertypes.CDKTFLanguage, kind providertypes.DocItemKind, name providertypes.DocItemName) indexstorage.Path {
	name = name.Normalize()
	if kind == providertypes.DocItemKindRoot {
		return indexstorage.Path(path.Join(providerAddr.Namespace, providerAddr.Name, string(version), cdktfDirName, string(language), string(name)+".md"))
	}
	return indexstorage.Path(path.Join(providerAddr.Namespace, providerAddr.Name, string(version), cdktfDirName, string(language), string(kind)+"s", string(name)+".md"))
}

func (s storage) GetProviderCDKTFDocItem(ctx context.Context, providerAddr providertypes.ProviderAddr, version string, language providertypes.CDKTFLanguage, kind providertypes.DocItemKind, name providertypes.DocItemName) ([]byte, error) {
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

func (s storage) StoreProviderCDKTFDocItem(ctx context.Context, providerAddr providertypes.ProviderAddr, version string, language providertypes.CDKTFLanguage, kind providertypes.DocItemKind, name providertypes.DocItemName, data []byte) error {
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
