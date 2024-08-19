package providerindexstorage

import (
	"github.com/opentofu/libregistry/types/provider"
)

type ProviderStoreFailedError struct {
	BaseError
	ProviderAddr provider.Addr
}

func (p *ProviderStoreFailedError) Error() string {
	if p.Cause != nil {
		return "Provider " + p.ProviderAddr.String() + " could not be stored: " + p.Cause.Error()
	}
	return "Provider " + p.ProviderAddr.String() + " could not be stored."
}
