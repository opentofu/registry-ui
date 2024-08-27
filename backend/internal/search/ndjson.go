package search

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"sync"
	"time"

	"github.com/opentofu/libregistry/logger"

	"github.com/opentofu/registry-ui/internal/search/searchstorage"
	"github.com/opentofu/registry-ui/internal/search/searchtypes"
)

func New(
	storage searchstorage.API,
	opts ...Opt,
) (API, error) {
	c := Config{}
	for _, opt := range opts {
		if err := opt(&c); err != nil {
			return nil, err
		}
	}

	if err := c.applyDefaults(); err != nil {
		return nil, err
	}

	return &api{
		cfg:     c,
		storage: storage,
		lock:    &sync.Mutex{},
	}, nil
}

type Config struct {
	Logger logger.Logger
}

func (c *Config) applyDefaults() error {
	if c.Logger == nil {
		c.Logger = logger.NewNoopLogger()
	}
	return nil
}

type Opt func(c *Config) error

func WithLogger(log logger.Logger) Opt {
	return func(c *Config) error {
		c.Logger = log
		return nil
	}
}

type api struct {
	metaIndex *searchtypes.MetaIndex
	storage   searchstorage.API
	lock      *sync.Mutex
	cfg       Config
}

func (a *api) RemoveVersionItems(ctx context.Context, itemType searchtypes.IndexType, addr string, version string) error {
	if err := a.ensureMetaIndex(ctx); err != nil {
		return err
	}

	return a.metaIndex.RemoveVersionItems(ctx, itemType, addr, version)
}

func (a *api) AddItem(ctx context.Context, item searchtypes.IndexItem) error {
	if err := a.ensureMetaIndex(ctx); err != nil {
		return err
	}
	item.LastUpdated = time.Now()
	return a.metaIndex.AddItem(ctx, item)
}

func (a *api) RemoveItem(ctx context.Context, id searchtypes.IndexID) error {
	if err := a.ensureMetaIndex(ctx); err != nil {
		return err
	}

	return a.metaIndex.RemoveItem(ctx, id)
}

func (a *api) GenerateIndex(ctx context.Context) error {
	if err := a.ensureMetaIndex(ctx); err != nil {
		return err
	}

	a.lock.Lock()
	defer a.lock.Unlock()

	buf := &bytes.Buffer{}
	writeItem := func(id string, item searchtypes.GeneratedIndexItem) error {
		itemJSON, err := json.Marshal(item)
		if err != nil {
			return fmt.Errorf("failed to create JSON for %s (%w)", id, err)
		}
		buf.Write(itemJSON)
		buf.Write([]byte("\n"))
		return nil
	}
	if err := writeItem("header", searchtypes.GeneratedIndexItem{
		Type: searchtypes.GeneratedIndexItemHeader,
		Header: &searchtypes.GeneratedIndexHeader{
			LastUpdated: time.Now(),
		},
	}); err != nil {
		return err
	}

	for i, t := range a.metaIndex.Deletions {
		generatedItem := searchtypes.GeneratedIndexItem{
			Type: searchtypes.GeneratedIndexItemDelete,
			Deletion: &searchtypes.ItemDeletion{
				ID:        i,
				DeletedAt: t,
			},
		}
		if err := writeItem(string(i), generatedItem); err != nil {
			return err
		}
	}

	for i, item := range a.metaIndex.Items {
		generatedItem := searchtypes.GeneratedIndexItem{
			Type:     searchtypes.GeneratedIndexItemAdd,
			Addition: &item,
		}
		if err := writeItem(string(i), generatedItem); err != nil {
			return err
		}
	}

	if err := a.storage.StoreMetaIndex(ctx, *a.metaIndex); err != nil {
		return err
	}
	if err := a.storage.StoreGeneratedIndex(ctx, buf.Bytes()); err != nil {
		return err
	}
	return nil
}

func (a *api) ensureMetaIndex(ctx context.Context) error {
	a.lock.Lock()
	if a.metaIndex == nil {
		metaIndex, err := a.storage.GetMetaIndex(ctx)
		if err != nil {
			var notFound *searchstorage.MetaIndexNotFoundError
			if !errors.As(err, &notFound) {
				a.lock.Unlock()
				return err
			}
			// Ignore not found errors, create a new meta index.
			// TODO do we need to signal back that the meta index is possibly incomplete? Do we need to make the
			//      error behavior configurable for full vs. partial runs?
		}
		a.metaIndex = &metaIndex
	}
	a.lock.Unlock()
	return nil
}
