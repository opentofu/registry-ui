package internal

import (
	"context"
	"fmt"
	"regexp"

	"github.com/opentofu/libregistry/logger"
	"github.com/opentofu/libregistry/types/module"
	"github.com/opentofu/libregistry/types/provider"
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
	NamespacePrefix     string
	Namespace           string
	Name                string
	TargetSystem        string
	ForceRepoDataUpdate bool
	ForceRegenerate     ForceRegenerateType
}

func WithForceRegenerateNamespace(namespace string) GenerateOpt {
	return func(c *GenerateConfig) error {
		if namespace == "" {
			return fmt.Errorf("empty namespace provided")
		}
		c.ForceRegenerate = append(c.ForceRegenerate, ForceRegenerateEntry{
			Namespace: namespace,
		})
		return nil
	}
}

func WithForceRegenerateNamespaceAndName(namespace string, name string) GenerateOpt {
	return func(c *GenerateConfig) error {
		if namespace == "" {
			return fmt.Errorf("empty namespace provided")
		}
		c.ForceRegenerate = append(c.ForceRegenerate, ForceRegenerateEntry{
			Namespace: namespace,
			Name:      name,
		})
		return nil
	}
}

func WithForceRegenerateSingleModule(addr module.Addr) GenerateOpt {
	return func(c *GenerateConfig) error {
		if err := addr.Validate(); err != nil {
			return err
		}
		c.ForceRegenerate = append(c.ForceRegenerate, ForceRegenerateEntry{
			Namespace:    addr.Namespace,
			Name:         addr.Name,
			TargetSystem: addr.TargetSystem,
		})
		return nil
	}
}

type ForceRegenerateType []ForceRegenerateEntry

func (f ForceRegenerateType) MustRegenerateModule(ctx context.Context, addr module.Addr) bool {
	for _, entry := range f {
		if entry.MustRegenerateModule(ctx, addr) {
			return true
		}
	}
	return false
}

func (f ForceRegenerateType) MustRegenerateProvider(ctx context.Context, addr provider.Addr) bool {
	for _, entry := range f {
		if entry.MustRegenerateProvider(ctx, addr) {
			return true
		}
	}
	return false
}

type ForceRegenerateEntry struct {
	Namespace    string
	Name         string
	TargetSystem string
}

func (f ForceRegenerateEntry) MustRegenerateModule(_ context.Context, addr module.Addr) bool {
	if f.Name == "" {
		return module.NormalizeNamespace(f.Namespace) == addr.Namespace
	}
	if f.TargetSystem == "" {
		return module.NormalizeNamespace(f.Namespace) == addr.Namespace && module.NormalizeName(f.Name) == addr.Name
	}
	return addr.Equals(module.Addr{
		Namespace:    f.Namespace,
		Name:         f.Name,
		TargetSystem: f.TargetSystem,
	})
}

func (f ForceRegenerateEntry) MustRegenerateProvider(_ context.Context, addr provider.Addr) bool {
	if f.TargetSystem != "" {
		return false
	}
	if f.Name == "" {
		return provider.NormalizeNamespace(f.Namespace) == addr.Namespace
	}
	return addr.Equals(provider.Addr{
		Namespace: f.Namespace,
		Name:      f.Name,
	})
}

func (g GenerateConfig) validate() error { //nolint:unused
	if g.Name != "" && g.Namespace == "" {
		return fmt.Errorf("cannot use name filtering without namespace filtering")
	}
	if g.TargetSystem != "" && g.Name == "" {
		return fmt.Errorf("cannot use target system filtering without name filtering")
	}
	if g.TargetSystem != "" && g.SkipUpdateModules {
		return fmt.Errorf("target system filtering makes no sense without module updates")
	}
	return nil
}

func WithSkipUpdateProviders(skip bool) GenerateOpt {
	return func(c *GenerateConfig) error {
		c.SkipUpdateProviders = skip
		return nil
	}
}

func WithForceRepoDataUpdate(force bool) GenerateOpt {
	return func(c *GenerateConfig) error {
		c.ForceRepoDataUpdate = force
		return nil
	}
}

func WithSkipUpdateModules(skip bool) GenerateOpt {
	return func(c *GenerateConfig) error {
		c.SkipUpdateModules = skip
		return nil
	}
}

var namespaceRe = regexp.MustCompile("^[a-zA-Z0-9._-]*$")
var nameRe = regexp.MustCompile("^[a-zA-Z0-9._-]*$")
var targetSystemRe = regexp.MustCompile("^[a-zA-Z0-9._-]*$")

func WithNamespacePrefix(namespacePrefix string) GenerateOpt {
	return func(c *GenerateConfig) error {
		if namespacePrefix != "" && c.Namespace != "" {
			return fmt.Errorf("filtering for both namespace and namespace prefix is not supported")
		}
		if !namespaceRe.MatchString(namespacePrefix) {
			return fmt.Errorf("invalid namespace: %s", namespaceRe)
		}
		c.NamespacePrefix = namespacePrefix
		return nil
	}
}

func WithNamespace(namespace string) GenerateOpt {
	return func(c *GenerateConfig) error {
		if namespace != "" && c.NamespacePrefix != "" {
			return fmt.Errorf("filtering for both namespace and namespace prefix is not supported")
		}
		if !namespaceRe.MatchString(namespace) {
			return fmt.Errorf("invalid namespace: %s", namespaceRe)
		}
		c.Namespace = namespace
		return nil
	}
}

func WithName(name string) GenerateOpt {
	return func(c *GenerateConfig) error {
		if !nameRe.MatchString(name) {
			return fmt.Errorf("invalid name: %s", nameRe)
		}
		c.Name = name
		return nil
	}
}

func WithTargetSystem(targetSystem string) GenerateOpt {
	return func(c *GenerateConfig) error {
		if !targetSystemRe.MatchString(targetSystem) {
			return fmt.Errorf("invalid target system: %s", targetSystemRe)
		}
		c.TargetSystem = targetSystem
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
		if cfg.TargetSystem != "" {
			if err := b.moduleIndexGenerator.GenerateSingleModule(ctx, module.Addr{
				Namespace:    cfg.Namespace,
				Name:         cfg.Name,
				TargetSystem: cfg.TargetSystem,
			}, moduleindex.WithForce(cfg.ForceRegenerate), moduleindex.WithForceRepoDataUpdate(cfg.ForceRepoDataUpdate)); err != nil {
				return fmt.Errorf("failed to generate modules (%w)", err)
			}
		} else if cfg.Name != "" {
			if err := b.moduleIndexGenerator.GenerateNamespaceAndName(ctx, cfg.Namespace, cfg.Name, moduleindex.WithForce(cfg.ForceRegenerate), moduleindex.WithForceRepoDataUpdate(cfg.ForceRepoDataUpdate)); err != nil {
				return fmt.Errorf("failed to generate modules (%w)", err)
			}
		} else if cfg.Namespace != "" {
			if err := b.moduleIndexGenerator.GenerateNamespace(ctx, cfg.Namespace, moduleindex.WithForce(cfg.ForceRegenerate), moduleindex.WithForceRepoDataUpdate(cfg.ForceRepoDataUpdate)); err != nil {
				return fmt.Errorf("failed to generate modules (%w)", err)
			}
		} else if cfg.NamespacePrefix != "" {
			if err := b.moduleIndexGenerator.GenerateNamespacePrefix(ctx, cfg.NamespacePrefix, moduleindex.WithForce(cfg.ForceRegenerate), moduleindex.WithForceRepoDataUpdate(cfg.ForceRepoDataUpdate)); err != nil {
				return fmt.Errorf("failed to generate modules (%w)", err)
			}
		} else {
			if err := b.moduleIndexGenerator.Generate(ctx, moduleindex.WithForce(cfg.ForceRegenerate), moduleindex.WithForceRepoDataUpdate(cfg.ForceRepoDataUpdate)); err != nil {
				return fmt.Errorf("failed to generate modules (%w)", err)
			}
		}
	}
	if !cfg.SkipUpdateProviders && cfg.TargetSystem == "" {
		if cfg.Name != "" {
			if err := b.providerIndexGenerator.GenerateSingleProvider(ctx, provider.Addr{Namespace: cfg.Namespace, Name: cfg.Name}, providerindex.WithForce(cfg.ForceRegenerate), providerindex.WithForceRepoDataUpdate(cfg.ForceRepoDataUpdate)); err != nil {
				return fmt.Errorf("failed to index providers (%w)", err)
			}
		} else if cfg.Namespace != "" {
			if err := b.providerIndexGenerator.GenerateNamespace(ctx, cfg.Namespace, providerindex.WithForce(cfg.ForceRegenerate), providerindex.WithForceRepoDataUpdate(cfg.ForceRepoDataUpdate)); err != nil {
				return fmt.Errorf("failed to index providers (%w)", err)
			}
		} else if cfg.NamespacePrefix != "" {
			if err := b.providerIndexGenerator.GenerateNamespacePrefix(ctx, cfg.NamespacePrefix, providerindex.WithForce(cfg.ForceRegenerate), providerindex.WithForceRepoDataUpdate(cfg.ForceRepoDataUpdate)); err != nil {
				return fmt.Errorf("failed to index providers (%w)", err)
			}
		} else {
			if err := b.providerIndexGenerator.Generate(ctx, providerindex.WithForce(cfg.ForceRegenerate), providerindex.WithForceRepoDataUpdate(cfg.ForceRepoDataUpdate)); err != nil {
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
