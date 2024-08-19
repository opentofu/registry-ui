package providerindexstorage

import (
	"github.com/opentofu/libregistry/types/provider"
)

type ProviderNotFoundError struct {
	BaseError

	ProviderAddr provider.Addr
}

func (p *ProviderNotFoundError) Error() string {
	if p.Cause != nil {
		return "Provider " + p.ProviderAddr.String() + " not found (" + p.Cause.Error() + ")."
	}
	return "Provider " + p.ProviderAddr.String() + " not found."
}
