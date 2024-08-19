package search_test

import (
	"context"
	"testing"

	"github.com/opentofu/libregistry/logger"
	"github.com/opentofu/registry-ui/internal/search"
	"github.com/opentofu/registry-ui/internal/search/searchstorage/filesystemsearchstorage"
	"github.com/opentofu/registry-ui/internal/search/searchtypes"
)

func TestLunrJSIndexing(t *testing.T) {
	item := searchtypes.IndexItem{
		ID:          "modules/test",
		Type:        searchtypes.IndexTypeModule,
		Title:       "test",
		Description: "",
		LinkVariables: map[string]string{
			"id": "test",
		},
		ParentID: "",
	}

	ctx := context.Background()

	storage, err := filesystemsearchstorage.New(t.TempDir())
	if err != nil {
		t.Fatal(err)
	}

	lunr, err := search.New(storage, search.WithLogger(logger.NewTestLogger(t)))
	if err != nil {
		t.Fatal(err)
	}

	if err := lunr.AddItem(ctx, item); err != nil {
		t.Fatal(err)
	}

	if err := lunr.GenerateIndex(ctx); err != nil {
		t.Fatal(err)
	}
}
