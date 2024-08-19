package moduleschema

import (
	"context"
)

// Extractor is a utility to extract module schemas.
type Extractor interface {
	// Extract extracts the module schema of a module present in the given directory.
	Extract(ctx context.Context, moduleDirectory string) (Schema, error)
}
