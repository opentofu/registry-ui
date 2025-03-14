package providerindexstorage

import "github.com/opentofu/registry-ui/internal/providerindex/providertypes"

type ProviderReadFailedError struct {
	BaseError
	ProviderAddr providertypes.ProviderAddr
}

func (p *ProviderReadFailedError) Error() string {
	if p.Cause != nil {
		return "Provider " + p.ProviderAddr.String() + " could not be read: " + p.Cause.Error()
	}
	return "Provider " + p.ProviderAddr.String() + " could not be read."
}
