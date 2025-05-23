package providertypes

import (
	"time"

	"github.com/opentofu/libregistry/types/provider"
)

// ProviderVersionDescriptor describes a provider version.
type ProviderVersionDescriptor struct {
	ID        provider.VersionNumber `json:"id"`
	Published time.Time              `json:"published"`
}
