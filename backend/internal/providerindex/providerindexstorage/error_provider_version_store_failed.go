package providerindexstorage

import (
	"github.com/opentofu/libregistry/types/provider"
)

type ProviderVersionStoreFailedError struct {
	BaseError
	ProviderAddr provider.Addr
	Version      provider.VersionNumber
}

func (p *ProviderVersionStoreFailedError) Error() string {
	if p.Cause != nil {
		return "Provider " + p.ProviderAddr.String() + " version " + string(p.Version) + " could not be stored: " + p.Cause.Error()
	}
	return "Provider " + p.ProviderAddr.String() + " version " + string(p.Version) + " could not be stored."
}
