package providerindex

import (
	"context"

	"github.com/opentofu/libregistry/types/provider"
)

// ForceRegenerate describes an interface to force regenerating a module even if it already exists in the index.
type ForceRegenerate interface {
	// MustRegenerateProvider returns true if a provider addr should be regenerated regardless of freshness state.
	// This function purposefully doesn't implement a by-version selection because it will interfere with the
	// generation of the search index.
	MustRegenerateProvider(ctx context.Context, addr provider.Addr) bool
}
