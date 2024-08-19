package providerindexstorage

import (
	"github.com/opentofu/libregistry/types/provider"
)

type ProviderVersionNotFoundError struct {
	BaseError

	ProviderAddr provider.Addr
	Version      provider.VersionNumber
}

func (p *ProviderVersionNotFoundError) Error() string {
	if p.Cause != nil {
		return "Provider " + p.ProviderAddr.String() + " version " + string(p.Version) + " not found (" + p.Cause.Error() + ")."
	}
	return "Provider " + p.ProviderAddr.String() + " version " + string(p.Version) + " not found."
}
