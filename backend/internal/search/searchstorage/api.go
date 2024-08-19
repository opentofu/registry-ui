package searchstorage

import (
	"context"

	"github.com/opentofu/registry-ui/internal/search/searchtypes"
)

type API interface {
	// GetMetaIndex returns the meta index from storage. If the meta index does not exist, it returns a preconfigured
	// empty searchtypes.MetaIndex and a *MetaIndexNotFoundError.
	GetMetaIndex(ctx context.Context) (searchtypes.MetaIndex, error)
	StoreMetaIndex(ctx context.Context, metaIndex searchtypes.MetaIndex) error
	StoreGeneratedIndex(ctx context.Context, data []byte) error
}

type MetaIndexNotFoundError struct {
	Cause error
}

func (m MetaIndexNotFoundError) Error() string {
	if m.Cause != nil {
		return "Meta index not found (" + m.Cause.Error() + ")"
	}
	return "Meta index not found"
}

func (m MetaIndexNotFoundError) Unwrap() error {
	return m.Cause
}
