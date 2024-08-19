package indexstoragesearch

import (
	"context"
	"encoding/json"
	"fmt"
	"os"

	"github.com/opentofu/registry-ui/internal/indexstorage"
	"github.com/opentofu/registry-ui/internal/search/searchstorage"
	"github.com/opentofu/registry-ui/internal/search/searchtypes"
)

func New(storage indexstorage.API) (searchstorage.API, error) {
	return &api{
		storage: storage,
	}, nil
}

type api struct {
	storage indexstorage.API
}

func (a api) GetMetaIndex(ctx context.Context) (searchtypes.MetaIndex, error) {
	metaIndex := searchtypes.NewMetaIndex()
	contents, err := a.storage.ReadFile(ctx, a.getMetaIndexFile())
	if err != nil {
		if os.IsNotExist(err) {
			return metaIndex, &searchstorage.MetaIndexNotFoundError{
				Cause: err,
			}
		}
		return metaIndex, err
	}
	if err := json.Unmarshal(contents, &metaIndex); err != nil {
		return metaIndex, fmt.Errorf("meta index corrupt (%w)", err)
	}

	return metaIndex, nil
}

func (a api) StoreMetaIndex(ctx context.Context, metaIndex searchtypes.MetaIndex) error {
	marshalled, err := json.Marshal(metaIndex)
	if err != nil {
		return fmt.Errorf("failed to write meta index (%w)", err)
	}
	return a.storage.WriteFile(ctx, a.getMetaIndexFile(), marshalled)
}

func (a api) StoreGeneratedIndex(ctx context.Context, data []byte) error {
	fileName := a.getGeneratedFile()
	if err := a.storage.WriteFile(ctx, fileName, data); err != nil {
		return fmt.Errorf("failed to write generated search index %s (%w)", fileName, err)
	}
	return nil
}

func (a api) getMetaIndexFile() indexstorage.Path {
	return "metaindex.json"
}

func (a api) getGeneratedFile() indexstorage.Path {
	return "search.json"
}
