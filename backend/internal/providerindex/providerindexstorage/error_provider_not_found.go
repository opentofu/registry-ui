package providerindexstorage

import "github.com/opentofu/registry-ui/internal/providerindex/providertypes"

type ProviderNotFoundError struct {
	BaseError

	ProviderAddr providertypes.ProviderAddr
}

func (p *ProviderNotFoundError) Error() string {
	if p.Cause != nil {
		return "Provider " + p.ProviderAddr.String() + " not found (" + p.Cause.Error() + ")."
	}
	return "Provider " + p.ProviderAddr.String() + " not found."
}
