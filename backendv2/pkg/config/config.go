package config

import (
	"context"
	"fmt"
	"log/slog"
	"os"
	"strings"

	"github.com/knadh/koanf/parsers/yaml"
	"github.com/knadh/koanf/providers/env/v2"
	"github.com/knadh/koanf/providers/file"
	"github.com/knadh/koanf/v2"
)

var k = koanf.New(".")

const EnvVarPrefix = "REGISTRY_"

type BackendConfig struct {
	Bucket      BucketConfig      `koanf:"bucket"`
	Telemetry   TelemetryConfig   `koanf:"telemetry"`
	DB          DBConfig          `koanf:"db"`
	License     LicenseConfig     `koanf:"license"`
	Concurrency ConcurrencyConfig `koanf:"concurrency"`
	GitHub      GitHubConfig      `koanf:"github"`

	WorkDir      string `koanf:"workdir"`
	RegistryPath string `koanf:"registrypath"`
}

func (c *BackendConfig) Validate() error {
	if err := c.Bucket.Validate(); err != nil {
		return err
	}

	if err := c.DB.Validate(); err != nil {
		return err
	}

	if c.WorkDir == "" {
		return fmt.Errorf("workDir is required")
	}

	if c.RegistryPath == "" {
		return fmt.Errorf("registryPath is required")
	}

	// make the directory if it doesn't exist
	err := os.MkdirAll(c.WorkDir, 0o755)
	if err != nil {
		return fmt.Errorf("could not create workDir: %w", err)
	}

	err = os.MkdirAll(c.RegistryPath, 0o755)
	if err != nil {
		return fmt.Errorf("could not create registryPath: %w", err)
	}

	if err := c.License.Validate(); err != nil {
		return err
	}

	if err := c.Concurrency.Validate(); err != nil {
		return err
	}

	if err := c.GitHub.Validate(); err != nil {
		return err
	}

	return nil
}

const defaultConfigFile = "config.yaml"

func LoadConfig(ctx context.Context) (*BackendConfig, error) {
	slog.InfoContext(ctx, "Loading configuration")

	// Load yaml config.
	if err := k.Load(file.Provider(defaultConfigFile), yaml.Parser()); err != nil {
		slog.WarnContext(ctx, "Failed to load config file, using env vars only", "error", err, "filename", defaultConfigFile)
	} else {
		slog.InfoContext(ctx, "Loaded configuration from config file", "filename", defaultConfigFile)
	}

	// also load from env vars
	err := k.Load(env.Provider(".", env.Opt{
		Prefix: EnvVarPrefix,
		TransformFunc: func(k, v string) (string, any) {
			k = strings.ReplaceAll(strings.ToLower(strings.TrimPrefix(k, EnvVarPrefix)), "_", ".")
			return k, v
		},
	}), nil)
	if err != nil {
		return nil, err
	}
	slog.InfoContext(ctx, "Loaded environment variables with prefix", "prefix", EnvVarPrefix)

	var backendConfig BackendConfig
	if err := k.Unmarshal("", &backendConfig); err != nil {
		slog.ErrorContext(ctx, "Failed to unmarshal configuration", "error", err)
		return nil, err
	}

	if err := backendConfig.Validate(); err != nil {
		slog.ErrorContext(ctx, "Configuration validation failed", "error", err)
		return nil, err
	}

	slog.InfoContext(ctx, "Configuration loaded and validated successfully")
	return &backendConfig, nil
}
