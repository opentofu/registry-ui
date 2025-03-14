package providerindexstorage

import "github.com/opentofu/registry-ui/internal/providerindex/providertypes"

type ProviderVersionStoreFailedError struct {
	BaseError
	ProviderAddr providertypes.ProviderAddr
	Version      string
}

func (p *ProviderVersionStoreFailedError) Error() string {
	if p.Cause != nil {
		return "Provider " + p.ProviderAddr.String() + " version " + string(p.Version) + " could not be stored: " + p.Cause.Error()
	}
	return "Provider " + p.ProviderAddr.String() + " version " + string(p.Version) + " could not be stored."
}
