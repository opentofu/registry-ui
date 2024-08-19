package search

import (
	"context"

	"github.com/opentofu/registry-ui/internal/search/searchtypes"
)

type API interface {
	// AddItem queues the addition of an index item.
	AddItem(ctx context.Context, item searchtypes.IndexItem) error

	// RemoveVersionItems removes all items of a specific type matching a version.
	RemoveVersionItems(ctx context.Context, itemType searchtypes.IndexType, addr string, version string) error

	// RemoveItem removes an item with the specific ID and all items referencing this item as a parent.
	RemoveItem(ctx context.Context, id searchtypes.IndexID) error

	// GenerateIndex generates a search index with the items currently in the searchtypes.MetaIndex. This function
	// returns an opaque blob that should be passed to the frontend for use as a search index. (Under the hood this is
	// a lunr.js search index.)
	GenerateIndex(ctx context.Context) error
}
