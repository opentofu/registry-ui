package providerindexstorage

import "github.com/opentofu/registry-ui/internal/providerindex/providertypes"

type ProviderVersionReadFailedError struct {
	BaseError
	ProviderAddr providertypes.ProviderAddr
	Version      string
}

func (p *ProviderVersionReadFailedError) Error() string {
	if p.Cause != nil {
		return "Provider " + p.ProviderAddr.String() + " version " + string(p.Version) + " could not be read: " + p.Cause.Error()
	}
	return "Provider " + p.ProviderAddr.String() + " version " + string(p.Version) + " could not be read."
}
