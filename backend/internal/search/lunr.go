package search

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"os/exec"
	"path"
	"sync"

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
	NodePath string
	JSDir    string
	Logger   logger.Logger
}

func (c *Config) applyDefaults() error {
	if c.NodePath == "" {
		c.NodePath = defaultNodePath
		// TODO check if node is runnable
	}
	if c.JSDir == "" {
		tempDir, err := os.MkdirTemp(os.TempDir(), "search-")
		if err != nil {
			return fmt.Errorf("failed to create temporary directory (%w)", err)
		}
		c.JSDir = tempDir
	}
	if c.Logger == nil {
		c.Logger = logger.NewNoopLogger()
	}
	return nil
}

type Opt func(c *Config) error

func WithJSDir(dir string) Opt {
	return func(c *Config) error {
		c.JSDir = dir
		return nil
	}
}

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

	input := make([]searchtypes.IndexItem, len(a.metaIndex.Items))
	i := 0
	for _, item := range a.metaIndex.Items {
		input[i] = item
		i++
	}

	lunrInput, err := json.Marshal(input)
	if err != nil {
		return fmt.Errorf("failed to create lunr input (%w)", err)
	}

	// Eject lunr.js:
	lunrFile := path.Join(a.cfg.JSDir, "lunr.js")
	if err := os.WriteFile(lunrFile, []byte(lunrJSCode), 0600); err != nil {
		return fmt.Errorf("failed to write lunr.js code to %s (%w)", lunrFile, err)
	}
	defer func() {
		_ = os.RemoveAll(lunrFile)
	}()

	// Eject the add code:
	generateFile := path.Join(a.cfg.JSDir, "generate.js")
	if err := os.WriteFile(generateFile, addScript, 0600); err != nil {
		return fmt.Errorf("failed to write generate.js code to %s (%w)", generateFile, err)
	}
	defer func() {
		_ = os.RemoveAll(lunrFile)
	}()

	stdin := &bytes.Buffer{}
	stdout := &bytes.Buffer{}

	_, _ = stdin.Write(lunrInput)

	cmd := exec.CommandContext(ctx, a.cfg.NodePath, generateFile)
	cmd.Dir = a.cfg.JSDir
	cmd.Stdout = stdout
	cmd.Stdin = stdin
	cmd.Stderr = logger.NewWriter(ctx, a.cfg.Logger, logger.LevelWarning, a.cfg.NodePath+" "+generateFile+": ")
	if err := cmd.Run(); err != nil {
		var exitErr *exec.ExitError
		if errors.As(err, &exitErr) {
			if exitErr.ExitCode() != 0 {
				return fmt.Errorf("node exited with a non-zero exit code (%d)", exitErr.ExitCode())
			}
		} else {
			return err
		}
	}
	if err := a.storage.StoreMetaIndex(ctx, *a.metaIndex); err != nil {
		return err
	}
	if err := a.storage.StoreGeneratedIndex(ctx, stdout.Bytes()); err != nil {
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
