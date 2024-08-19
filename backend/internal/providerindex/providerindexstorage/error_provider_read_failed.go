package providerindexstorage

import (
	"github.com/opentofu/libregistry/types/provider"
)

type ProviderReadFailedError struct {
	BaseError
	ProviderAddr provider.Addr
}

func (p *ProviderReadFailedError) Error() string {
	if p.Cause != nil {
		return "Provider " + p.ProviderAddr.String() + " could not be read: " + p.Cause.Error()
	}
	return "Provider " + p.ProviderAddr.String() + " could not be read."
}
