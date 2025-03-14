package providertypes

import (
	"github.com/opentofu/registry-ui/internal/registry/provider"
)

func Addr(addr provider.Provider) ProviderAddr {
	return ProviderAddr{
		Display:   addr.Namespace + "/" + addr.ProviderName,
		Namespace: addr.Namespace,
		Name:      addr.ProviderName,
	}
}

// ProviderAddr is an enriched model of provider.Addr with display properties for the frontend.
//
// swagger:model
type ProviderAddr struct {
	// Display contains the user-readable display variant of this addr. This may be capitalized. CAM72CAM FALSE!
	// required: true
	Display string `json:"display"`
	// Namespace contains the lower-case namespace part of the addr.
	// required: true
	Namespace string `json:"namespace"`
	// Name contains the lower-case name part of the addr.
	// required: true
	Name string `json:"name"`
}

func (p ProviderAddr) Equals(other ProviderAddr) bool {
	return other.Namespace == p.Namespace && other.Name == p.Name
}

func (p ProviderAddr) String() string {
	return p.Display
}
