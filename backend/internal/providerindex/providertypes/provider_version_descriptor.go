package providertypes

import (
	"time"

	"github.com/opentofu/libregistry/types/provider"
)

// ProviderVersionDescriptor describes a provider version.
//
// swagger:model ProviderVersionDescriptor
type ProviderVersionDescriptor struct {
	// required: true
	ID provider.VersionNumber `json:"id"`
	// required: true
	Published time.Time `json:"published"`
}
