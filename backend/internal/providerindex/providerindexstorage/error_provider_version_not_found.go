package providerindexstorage

import "github.com/opentofu/registry-ui/internal/providerindex/providertypes"

type ProviderVersionNotFoundError struct {
	BaseError

	ProviderAddr providertypes.ProviderAddr
	Version      string
}

func (p *ProviderVersionNotFoundError) Error() string {
	if p.Cause != nil {
		return "Provider " + p.ProviderAddr.String() + " version " + string(p.Version) + " not found (" + p.Cause.Error() + ")."
	}
	return "Provider " + p.ProviderAddr.String() + " version " + string(p.Version) + " not found."
}
