package internal

import (
	"context"
	"fmt"
	"regexp"

	"github.com/opentofu/libregistry/logger"
	"golang.org/x/sync/errgroup"

	"github.com/opentofu/registry-ui/internal/indexstorage"
	"github.com/opentofu/registry-ui/internal/moduleindex"
	"github.com/opentofu/registry-ui/internal/providerindex"
	"github.com/opentofu/registry-ui/internal/registrycloner"
	"github.com/opentofu/registry-ui/internal/search"
	"github.com/opentofu/registry-ui/internal/server"
)

func New(
	cloner registrycloner.Cloner,
	moduleIndexGenerator moduleindex.Generator,
	providerIndexGenerator providerindex.DocumentationGenerator,
	searchAPI search.API,
	openAPIWriter server.OpenAPIWriter,
	storage indexstorage.Committable,
	opts ...Opt,
) (Backend, error) {
	cfg := Config{}
	for _, opt := range opts {
		if err := opt(&cfg); err != nil {
			return nil, err
		}
	}
	if err := cfg.ApplyDefaults(); err != nil {
		return nil, err
	}
	return &backend{
		cloner:                 cloner,
		moduleIndexGenerator:   moduleIndexGenerator,
		providerIndexGenerator: providerIndexGenerator,
		searchAPI:              searchAPI,
		openAPIWriter:          openAPIWriter,
		cfg:                    cfg,
		storage:                storage,
	}, nil
}

type Backend interface {
	Generate(ctx context.Context, opts ...GenerateOpt) error
}

type Opt func(c *Config) error

type Config struct {
	Logger logger.Logger
}

func (c *Config) ApplyDefaults() error {
	if c.Logger == nil {
		c.Logger = logger.NewNoopLogger()
	}
	c.Logger = c.Logger.WithName("Backend")
	return nil
}

func WithLogger(log logger.Logger) Opt {
	return func(c *Config) error {
		c.Logger = log
		return nil
	}
}

type GenerateOpt func(c *GenerateConfig) error

type GenerateConfig struct {
	SkipUpdateProviders bool
	SkipUpdateModules   bool
	Namespace           string
}

func WithSkipUpdateProviders(skip bool) GenerateOpt {
	return func(c *GenerateConfig) error {
		c.SkipUpdateProviders = skip
		if c.SkipUpdateProviders && c.SkipUpdateModules {
			return fmt.Errorf("skipping both provider and module updates results in a noop generation")
		}
		return nil
	}
}

func WithSkipUpdateModules(skip bool) GenerateOpt {
	return func(c *GenerateConfig) error {
		c.SkipUpdateModules = skip
		if c.SkipUpdateProviders && c.SkipUpdateModules {
			return fmt.Errorf("skipping both provider and module updates results in a noop generation")
		}
		return nil
	}
}

var namespaceRe = regexp.MustCompile("^[a-zA-Z0-9._-]*$")

func WithNamespace(namespace string) GenerateOpt {
	return func(c *GenerateConfig) error {
		if !namespaceRe.MatchString(namespace) {
			return fmt.Errorf("invalid namespace: %s", namespaceRe)
		}
		c.Namespace = namespace
		return nil
	}
}

type backend struct {
	cfg                    Config
	moduleIndexGenerator   moduleindex.Generator
	providerIndexGenerator providerindex.DocumentationGenerator
	searchAPI              search.API
	cloner                 registrycloner.Cloner
	openAPIWriter          server.OpenAPIWriter
	storage                indexstorage.Committable
}

func (b backend) Generate(ctx context.Context, opts ...GenerateOpt) error {
	cfg := GenerateConfig{}
	for _, opt := range opts {
		if err := opt(&cfg); err != nil {
			return err
		}
	}

	eg := &errgroup.Group{}
	eg.Go(func() error {
		// TODO only do this if a commit has previously started, otherwise roll back. This needs a change in the
		//      data structure marking that a commit has started.
		b.cfg.Logger.Info(ctx, "Committing any previous changes...")
		// Commit changes from a previous failed commit before we begin so we have a clean directory and can safely
		// roll back if generate fails.
		if err := b.storage.Recover(ctx); err != nil {
			b.cfg.Logger.Error(ctx, "Commit failed (%v).", err)
			return err
		}
		b.cfg.Logger.Info(ctx, "Commit complete.")
		return nil
	})
	eg.Go(func() error {
		b.cfg.Logger.Info(ctx, "Cloning registry repository...")
		if err := b.cloner.Clone(ctx); err != nil {
			b.cfg.Logger.Error(ctx, "Clone failed (%v).", err)
			return fmt.Errorf("failed to clone registry (%w)", err)
		}
		b.cfg.Logger.Info(ctx, "Clone complete.")
		return nil
	})
	if err := eg.Wait(); err != nil {
		return err
	}
	b.cfg.Logger.Info(ctx, "Generating artifacts...")
	err := b.generate(ctx, cfg)
	if err != nil {
		b.cfg.Logger.Warn(ctx, "Generation failed, rolling back changes (%v)...", err)
		if rollbackErr := b.storage.Rollback(ctx); rollbackErr != nil {
			b.cfg.Logger.Error(ctx, "Rollback failed (%v).", err)
			return fmt.Errorf("rollback failed (%s) on generate failure (%w)", rollbackErr, err)
		}
		b.cfg.Logger.Info(ctx, "Rollback complete.")
		return err
	}
	b.cfg.Logger.Info(ctx, "Generation complete, committing changes...")
	if err := b.storage.Commit(ctx); err != nil {
		b.cfg.Logger.Warn(ctx, "Commit failed (%v). Please save the storage directory and re-run commit.", err)
		return err
	}
	b.cfg.Logger.Info(ctx, "Commit complete.")
	return nil
}

func (b backend) generate(ctx context.Context, cfg GenerateConfig) error {
	if !cfg.SkipUpdateModules {
		if cfg.Namespace != "" {
			if err := b.moduleIndexGenerator.GenerateNamespace(ctx, cfg.Namespace); err != nil {
				return fmt.Errorf("failed to generate modules (%w)", err)
			}
		} else {
			if err := b.moduleIndexGenerator.Generate(ctx); err != nil {
				return fmt.Errorf("failed to generate modules (%w)", err)
			}
		}
	}
	if !cfg.SkipUpdateProviders {
		if cfg.Namespace != "" {
			if err := b.providerIndexGenerator.GenerateNamespace(ctx, cfg.Namespace); err != nil {
				return fmt.Errorf("failed to index providers (%w)", err)
			}
		} else {
			if err := b.providerIndexGenerator.Generate(ctx); err != nil {
				return fmt.Errorf("failed to index providers (%w)", err)
			}
		}
	}

	if err := b.searchAPI.GenerateIndex(ctx); err != nil {
		return fmt.Errorf("failed to update search index (%w)", err)
	}

	if err := b.openAPIWriter.Write(ctx); err != nil {
		return fmt.Errorf("failed to write OpenAPI spec (%w)", err)
	}
	return nil
}
