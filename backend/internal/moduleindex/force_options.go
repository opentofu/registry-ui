package moduleindex

import (
	"context"

	"github.com/opentofu/libregistry/types/module"
)

// ForceRegenerate describes an interface to force regenerating a module even if it already exists in the index.
type ForceRegenerate interface {
	// MustRegenerateModule returns true if a module addr should be regenerated regardless of freshness state.
	// This function purposefully doesn't implement a by-version selection because it will interfere with the
	// generation of the search index.
	MustRegenerateModule(ctx context.Context, addr module.Addr) bool
}
