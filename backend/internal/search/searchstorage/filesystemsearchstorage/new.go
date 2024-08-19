package filesystemsearchstorage

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"path"

	"github.com/opentofu/registry-ui/internal/search/searchstorage"
	"github.com/opentofu/registry-ui/internal/search/searchtypes"
)

func New(dir string) (searchstorage.API, error) {
	stat, err := os.Stat(dir)
	if err != nil {
		return nil, fmt.Errorf("directory %s not found (%w)", dir, err)
	}
	if !stat.IsDir() {
		return nil, fmt.Errorf("%s is not a directory", dir)
	}
	return &filesystem{
		dir: dir,
	}, nil
}

type filesystem struct {
	dir string
}

func (f filesystem) StoreGeneratedIndex(_ context.Context, data []byte) error {
	fileName := f.getGeneratedFile()
	if err := os.WriteFile(fileName, data, 0644); err != nil {
		return fmt.Errorf("failed to write generated search index %s (%w)", fileName, err)
	}
	return nil
}

func (f filesystem) GetMetaIndex(_ context.Context) (searchtypes.MetaIndex, error) {
	metaIndex := searchtypes.NewMetaIndex()
	contents, err := os.ReadFile(f.getMetaIndexFile())
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

func (f filesystem) StoreMetaIndex(_ context.Context, metaIndex searchtypes.MetaIndex) error {
	marshalled, err := json.Marshal(metaIndex)
	if err != nil {
		return fmt.Errorf("failed to write meta index (%w)", err)
	}
	return os.WriteFile(f.getMetaIndexFile(), marshalled, 0644)
}

func (f filesystem) getMetaIndexFile() string {
	return path.Join(f.dir, "metaindex.json")
}

func (f filesystem) getGeneratedFile() string {
	return path.Join(f.dir, "search.json")
}
