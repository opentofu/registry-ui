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
	"go.opentelemetry.io/otel"
)

var k = koanf.New(".")

type BackendConfig struct {
	Bucket      BucketConfig      `koanf:"bucket"`
	Telemetry   TelemetryConfig   `koanf:"telemetry"`
	DB          DBConfig          `koanf:"db"`
	License     LicenseConfig     `koanf:"license"`
	Concurrency ConcurrencyConfig `koanf:"concurrency"`

	WorkDir      string `koanf:"workDir"`
	RegistryPath string `koanf:"registryPath"`
}

func (c *BackendConfig) validate() error {
	if err := c.Bucket.validate(); err != nil {
		return err
	}

	if err := c.DB.validate(); err != nil {
		return err
	}

	// no need to validate telemetry right noww

	if c.WorkDir == "" {
		return fmt.Errorf("workDir is required")
	}

	if c.RegistryPath == "" {
		return fmt.Errorf("registryPath is required")
	}

	// make the directory if it doesnt eixst
	err := os.MkdirAll(c.WorkDir, 0755)
	if err != nil {
		return fmt.Errorf("could not create workDir: %w", err)
	}

	err = os.MkdirAll(c.RegistryPath, 0755)
	if err != nil {
		return fmt.Errorf("could not create registryPath: %w", err)
	}

	if err := c.License.Validate(); err != nil {
		return err
	}

	if err := c.Concurrency.Validate(); err != nil {
		return err
	}
	return nil
}

const EnvVarPrefix = "REGISTRY_"

func LoadConfig() (*BackendConfig, error) {
	ctx := context.Background()
	tracer := otel.Tracer("opentofu-registry-backend")
	ctx, span := tracer.Start(ctx, "LoadConfig")
	defer span.End()

	slog.InfoContext(ctx, "Loading configuration")

	// Load yaml config.
	if err := k.Load(file.Provider("config.yaml"), yaml.Parser()); err != nil {
		slog.WarnContext(ctx, "Failed to load config.yaml, using env vars only", "error", err)
	} else {
		slog.InfoContext(ctx, "Loaded configuration from config.yaml")
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

	if err := backendConfig.validate(); err != nil {
		slog.ErrorContext(ctx, "Configuration validation failed", "error", err)
		return nil, err
	}

	slog.InfoContext(ctx, "Configuration loaded and validated successfully")
	return &backendConfig, nil
}
