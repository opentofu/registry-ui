package providerdocsource

import (
	"context"

	"github.com/opentofu/libregistry/types/provider"
	"github.com/opentofu/registry-ui/internal/license"
	"github.com/opentofu/registry-ui/internal/providerindex/providerindexstorage"
	"github.com/opentofu/registry-ui/internal/providerindex/providertypes"
)

type documentation struct {
	docs     *providerDoc
	cdktf    map[string]Documentation
	licenses license.List
	link     string
}

func (d documentation) Store(ctx context.Context, addr provider.Addr, version providertypes.ProviderVersionDescriptor, storage providerindexstorage.API) (providertypes.ProviderVersion, error) {
	versionData := d.ToProviderTypes(ctx, version)

	if err := storage.StoreProviderVersion(ctx, addr, versionData); err != nil {
		return versionData, err
	}

	if err := d.docs.Store(ctx, addr, version, storage, ""); err != nil {
		return versionData, err
	}
	for lang, doc := range d.cdktf {
		if err := doc.Store(ctx, addr, version, storage, providertypes.CDKTFLanguage(lang)); err != nil {
			return versionData, err
		}
	}
	return versionData, nil
}

func (d documentation) ToProviderTypes(ctx context.Context, version providertypes.ProviderVersionDescriptor) providertypes.ProviderVersion {
	return providertypes.ProviderVersion{
		ProviderVersionDescriptor: version,
		Docs:                      d.docs.ToProviderTypes(ctx),
		CDKTFDocs:                 map[providertypes.CDKTFLanguage]providertypes.ProviderDocs{},
		Licenses:                  d.licenses,
		Link:                      d.link,
	}
}

func (d documentation) GetDocumentation(_ context.Context) (Documentation, error) {
	return d.docs, nil
}

func (d documentation) GetCDKTF(_ context.Context) (map[string]Documentation, error) {
	return d.cdktf, nil
}

func (d documentation) GetLicenses(_ context.Context) (license.List, error) {
	return d.licenses, nil
}
