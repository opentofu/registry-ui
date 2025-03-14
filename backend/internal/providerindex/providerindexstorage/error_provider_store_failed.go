package providerindexstorage

import "github.com/opentofu/registry-ui/internal/providerindex/providertypes"

type ProviderStoreFailedError struct {
	BaseError
	ProviderAddr providertypes.ProviderAddr
}

func (p *ProviderStoreFailedError) Error() string {
	if p.Cause != nil {
		return "Provider " + p.ProviderAddr.String() + " could not be stored: " + p.Cause.Error()
	}
	return "Provider " + p.ProviderAddr.String() + " could not be stored."
}
