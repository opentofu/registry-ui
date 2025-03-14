package providertypes

import (
	"time"
)

// ProviderVersionDescriptor describes a provider version.
//
// swagger:model ProviderVersionDescriptor
type ProviderVersionDescriptor struct {
	// required: true
	ID string `json:"id"`
	// required: true
	Published time.Time `json:"published"`
}
