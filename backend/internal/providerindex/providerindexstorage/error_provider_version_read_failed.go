package providerindexstorage

import (
	"github.com/opentofu/libregistry/types/provider"
)

type ProviderVersionReadFailedError struct {
	BaseError
	ProviderAddr provider.Addr
	Version      provider.VersionNumber
}

func (p *ProviderVersionReadFailedError) Error() string {
	if p.Cause != nil {
		return "Provider " + p.ProviderAddr.String() + " version " + string(p.Version) + " could not be read: " + p.Cause.Error()
	}
	return "Provider " + p.ProviderAddr.String() + " version " + string(p.Version) + " could not be read."
}
