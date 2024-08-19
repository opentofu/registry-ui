package registrycloner

import (
	"context"
	"fmt"
	"os"
	"path"
	"path/filepath"
	"regexp"

	"github.com/opentofu/registry-ui/internal/defaults"
)

func New(opts ...Opt) (Cloner, error) {
	cfg := Config{}
	for _, opt := range opts {
		if err := opt(&cfg); err != nil {
			return nil, err
		}
	}
	if err := cfg.applyDefaults(); err != nil {
		return nil, err
	}
	return &cloner{
		cfg: cfg,
	}, nil
}

type Opt func(c *Config) error

type Config struct {
	Repo      string
	Directory string
	Ref       string
}

func (c *Config) applyDefaults() error {
	if c.Repo == "" {
		c.Repo = defaults.RegistryRepo
	}
	if c.Directory == "" {
		c.Directory = defaults.RegistryDir
	}
	if c.Ref == "" {
		c.Ref = defaults.RegistryRef
	}
	return nil
}

const maxRegistryRepoLength = 2048

var registryRepoRe = regexp.MustCompile("^[a-zA-Z0-9/:._-]+$")

func WithRepo(repo string) Opt {
	return func(c *Config) error {
		if len(repo) > maxRegistryRepoLength || !registryRepoRe.MatchString(repo) {
			return fmt.Errorf("invalid repository %s", repo)
		}
		c.Repo = repo
		return nil
	}
}

func WithDirectory(dir string) Opt {
	return func(c *Config) error {
		if dir == "" {
			return fmt.Errorf("the registry directory cannot be empty")
		}
		base := filepath.Base(dir)
		parent := filepath.Dir(dir)
		stat, err := os.Stat(parent)
		if err != nil {
			return fmt.Errorf("parent directory (%s) for registry does not exist (%w)", parent, err)
		}
		if !stat.IsDir() {
			return fmt.Errorf("parent directory (%s) for registry is not a directory", parent)
		}
		absParent, err := filepath.Abs(parent)
		if err != nil {
			return fmt.Errorf("failed to determine absolute path for registry parent dir %s (%w)", parent, err)
		}
		c.Directory = path.Join(absParent, base)
		return nil
	}
}

func WithRef(ref string) Opt {
	return func(c *Config) error {
		c.Ref = ref
		return nil
	}
}

type Cloner interface {
	Clone(ctx context.Context) error
}

type cloner struct {
	cfg Config
}

func (c cloner) Clone(ctx context.Context) error {
	if _, err := os.Stat(path.Join(c.cfg.Directory, ".git")); os.IsNotExist(err) {
		if err := c.cloneGitRepo(ctx); err != nil {
			return fmt.Errorf("failed to clone registry into %s (%w)", c.cfg.Directory, err)
		}
	}

	if err := c.clean(ctx); err != nil {
		return fmt.Errorf("error cleaning: %w", err)
	}

	if err := c.reset(ctx); err != nil {
		return fmt.Errorf("error resetting: %w", err)
	}

	if err := c.switchRef(ctx); err != nil {
		return fmt.Errorf("error switching to %s: %w", c.cfg.Ref, err)
	}

	if err := c.pullLatest(ctx); err != nil {
		return fmt.Errorf("error pulling latest changes: %w", err)
	}
	return nil
}
