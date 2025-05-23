package providertypes

import (
	"encoding/json"

	"github.com/opentofu/libregistry/types/provider"
)

func Addr(addr provider.Addr) ProviderAddr {
	return ProviderAddr{
		Addr:      addr,
		Display:   addr.String(),
		Namespace: addr.Namespace,
		Name:      addr.Name,
	}
}

// ProviderAddr is an enriched model of provider.Addr with display properties for the frontend.
type ProviderAddr struct {
	provider.Addr

	// Display contains the user-readable display variant of this addr. This may be capitalized.
	Display string `json:"display"`
	// Namespace contains the lower-case namespace part of the addr.
	Namespace string `json:"namespace"`
	// Name contains the lower-case name part of the addr.
	Name string `json:"name"`
}

type marshalledProviderAddr struct {
	Display   string `json:"display"`
	Namespace string `json:"namespace"`
	Name      string `json:"name"`
}

func (p *ProviderAddr) UnmarshalJSON(data []byte) error {
	marshalled := marshalledProviderAddr{}
	if err := json.Unmarshal(data, &marshalled); err != nil {
		return err
	}

	*p = ProviderAddr{
		Addr:      provider.Addr{Namespace: marshalled.Namespace, Name: marshalled.Name},
		Display:   marshalled.Display,
		Namespace: marshalled.Namespace,
		Name:      marshalled.Name,
	}
	return nil
}

func (p *ProviderAddr) MarshalJSON() ([]byte, error) {
	return json.Marshal(marshalledProviderAddr{
		Display:   p.Display,
		Namespace: p.Namespace,
		Name:      p.Name,
	})
}
