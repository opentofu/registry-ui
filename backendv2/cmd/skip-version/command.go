// Package skipversion implements the command to manually mark a provider or module version as skipped
package skipversion

import (
	"context"
	"fmt"
	"log/slog"

	"github.com/urfave/cli/v3"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/codes"

	"github.com/opentofu/registry-ui/pkg/config"
	"github.com/opentofu/registry-ui/pkg/module/storage"
	providerstorage "github.com/opentofu/registry-ui/pkg/provider/storage"
	"github.com/opentofu/registry-ui/pkg/telemetry"
)

func NewCommand(cfg *config.BackendConfig) *cli.Command {
	return &cli.Command{
		Name:  "skip-version",
		Usage: "Manually mark a provider or module version as skipped",
		Flags: []cli.Flag{
			&cli.StringFlag{
				Name:     "type",
				Aliases:  []string{"t"},
				Usage:    "Resource type: 'provider' or 'module'",
				Required: true,
			},
			&cli.StringFlag{
				Name:     "namespace",
				Aliases:  []string{"n"},
				Usage:    "Namespace (e.g., hashicorp)",
				Required: true,
			},
			&cli.StringFlag{
				Name:     "name",
				Usage:    "Name (e.g., aws for provider, vpc for module)",
				Required: true,
			},
			&cli.StringFlag{
				Name:     "target",
				Usage:    "Module target system (e.g., aws) - required for modules only",
				Required: false,
			},
			&cli.StringFlag{
				Name:     "version",
				Aliases:  []string{"v"},
				Usage:    "Version to skip (e.g., 1.2.3)",
				Required: true,
			},
			&cli.StringFlag{
				Name:     "reason",
				Aliases:  []string{"r"},
				Usage:    "Skip reason: incompatible_license, no_license, processing_error, manual_skip, malformed_data",
				Required: true,
			},
			&cli.StringFlag{
				Name:     "message",
				Aliases:  []string{"m"},
				Usage:    "Optional detailed message explaining why this version is being skipped",
				Required: false,
			},
		},
		Action: func(ctx context.Context, cmd *cli.Command) error {
			return run(ctx, cmd, cfg)
		},
	}
}

func run(ctx context.Context, cmd *cli.Command, cfg *config.BackendConfig) error {
	ctx, span := telemetry.Tracer().Start(ctx, "skip-version")
	defer span.End()

	resourceType := cmd.String("type")
	namespace := cmd.String("namespace")
	name := cmd.String("name")
	target := cmd.String("target")
	version := cmd.String("version")
	skipReason := cmd.String("reason")
	errorMessage := cmd.String("message")

	// Validate resource type
	if resourceType != "provider" && resourceType != "module" {
		err := fmt.Errorf("invalid type: %s (must be 'provider' or 'module')", resourceType)
		span.RecordError(err)
		span.SetStatus(codes.Error, err.Error())
		return err
	}

	// Validate target for modules
	if resourceType == "module" && target == "" {
		err := fmt.Errorf("--target is required for modules")
		span.RecordError(err)
		span.SetStatus(codes.Error, err.Error())
		return err
	}

	// Validate skip reason
	validReasons := map[string]bool{
		"incompatible_license": true,
		"no_license":           true,
		"processing_error":     true,
		"manual_skip":          true,
		"malformed_data":       true,
	}
	if !validReasons[skipReason] {
		err := fmt.Errorf("invalid reason: %s (must be one of: incompatible_license, no_license, processing_error, manual_skip, malformed_data)", skipReason)
		span.RecordError(err)
		span.SetStatus(codes.Error, err.Error())
		return err
	}

	span.SetAttributes(
		attribute.String("resource.type", resourceType),
		attribute.String("resource.namespace", namespace),
		attribute.String("resource.name", name),
		attribute.String("resource.version", version),
		attribute.String("skip.reason", skipReason),
	)

	if resourceType == "module" {
		span.SetAttributes(attribute.String("resource.target", target))
	}

	// Connect to database
	pool, err := cfg.DB.GetPool(ctx)
	if err != nil {
		span.RecordError(err)
		span.SetStatus(codes.Error, err.Error())
		slog.ErrorContext(ctx, "Failed to connect to database", "error", err)
		return fmt.Errorf("failed to connect to database: %w", err)
	}

	// Update the version status based on type
	if resourceType == "provider" {
		slog.InfoContext(ctx, "Marking provider version as skipped",
			"namespace", namespace, "name", name, "version", version, "reason", skipReason)

		err = providerstorage.UpdateProviderVersionStatus(ctx, pool, namespace, name, version, "skipped", skipReason, errorMessage)
		if err != nil {
			span.RecordError(err)
			span.SetStatus(codes.Error, err.Error())
			slog.ErrorContext(ctx, "Failed to mark provider version as skipped", "error", err)
			return fmt.Errorf("failed to mark provider version as skipped: %w", err)
		}

		slog.InfoContext(ctx, "Successfully marked provider version as skipped",
			"namespace", namespace, "name", name, "version", version)
		fmt.Printf("✓ Marked provider %s/%s@%s as skipped (reason: %s)\n", namespace, name, version, skipReason)
	} else {
		slog.InfoContext(ctx, "Marking module version as skipped",
			"namespace", namespace, "name", name, "target", target, "version", version, "reason", skipReason)

		err = storage.UpdateModuleVersionStatus(ctx, pool, namespace, name, target, version, "skipped", skipReason, errorMessage)
		if err != nil {
			span.RecordError(err)
			span.SetStatus(codes.Error, err.Error())
			slog.ErrorContext(ctx, "Failed to mark module version as skipped", "error", err)
			return fmt.Errorf("failed to mark module version as skipped: %w", err)
		}

		slog.InfoContext(ctx, "Successfully marked module version as skipped",
			"namespace", namespace, "name", name, "target", target, "version", version)
		fmt.Printf("✓ Marked module %s/%s/%s@%s as skipped (reason: %s)\n", namespace, name, target, version, skipReason)
	}

	return nil
}
