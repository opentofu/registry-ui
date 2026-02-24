// Package getmodulelicense implements the get-module-license command to test license detection for a specific module.
package getmodulelicense

import (
	"context"
	"fmt"
	"log/slog"

	"github.com/urfave/cli/v3"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/codes"

	"github.com/opentofu/registry-ui/pkg/config"
	"github.com/opentofu/registry-ui/pkg/module"
	"github.com/opentofu/registry-ui/pkg/telemetry"
)

func NewCommand() *cli.Command {
	return &cli.Command{
		Name:  "get-module-license",
		Usage: "Test license detection for a specific module",
		Flags: []cli.Flag{
			&cli.StringFlag{
				Name:     "namespace",
				Usage:    "Module namespace (e.g., hashicorp)",
				Required: true,
			},
			&cli.StringFlag{
				Name:     "name",
				Usage:    "Module name (e.g., vpc)",
				Required: true,
			},
			&cli.StringFlag{
				Name:     "target",
				Usage:    "Target provider/platform from the module address namespace/name/target (e.g., 'aws' in hashicorp/vpc/aws)",
				Required: true,
			},
			&cli.StringFlag{
				Name:     "version",
				Aliases:  []string{"v"},
				Usage:    "Module version (e.g., 3.0.0)",
				Required: true,
			},
		},
		Action: func(ctx context.Context, cmd *cli.Command) error {
			return run(ctx, cmd)
		},
	}
}

func run(ctx context.Context, cmd *cli.Command) error {
	cfg := config.FromCLI(cmd)
	ctx, span := telemetry.Tracer().Start(ctx, "cmd.get_module_license")
	defer span.End()

	namespace := cmd.String("namespace")
	name := cmd.String("name")
	target := cmd.String("target")
	version := cmd.String("version")

	span.SetAttributes(
		attribute.String("module.namespace", namespace),
		attribute.String("module.name", name),
		attribute.String("module.target", target),
		attribute.String("module.version", version),
	)

	slog.InfoContext(ctx, "License detection for module",
		"namespace", namespace, "name", name, "target", target, "version", version)

	// Create module reader
	moduleReader, err := module.NewModuleReader(ctx, cfg)
	if err != nil {
		span.RecordError(err)
		span.SetStatus(codes.Error, err.Error())
		slog.ErrorContext(ctx, "Failed to create module reader", "error", err)
		return fmt.Errorf("failed to create module reader: %w", err)
	}

	// Checkout the version for processing
	workDir, cleanup, err := moduleReader.CheckoutVersionForScraping(ctx, namespace, name, target, version)
	if err != nil {
		span.RecordError(err)
		span.SetStatus(codes.Error, err.Error())
		slog.ErrorContext(ctx, "Failed to checkout version", "error", err)
		return fmt.Errorf("failed to checkout version: %w", err)
	}
	defer cleanup()

	slog.InfoContext(ctx, "Checked out module version", "workDir", workDir)

	// Detect licenses in the directory
	licenses, err := moduleReader.DetectLicensesInDirectory(ctx, namespace, name, target, workDir)
	if err != nil {
		span.RecordError(err)
		span.SetStatus(codes.Error, err.Error())
		slog.ErrorContext(ctx, "Failed to detect licenses", "error", err)
		return fmt.Errorf("failed to detect licenses: %w", err)
	}

	// Display results
	fmt.Printf("\n=== License Detection Results ===\n")
	fmt.Printf("Module: %s/%s/%s@%s\n", namespace, name, target, version)
	fmt.Printf("Working Directory: %s\n", workDir)
	fmt.Printf("Licenses Found: %d\n\n", len(licenses))

	if len(licenses) == 0 {
		fmt.Println("No licenses detected.")
		return nil
	}

	// Pretty print each license
	for i, license := range licenses {
		fmt.Printf("License %d:\n", i+1)
		fmt.Printf("  SPDX: %s\n", license.SPDX)
		fmt.Printf("  File: %s\n", license.File)
		fmt.Printf("  Confidence: %.4f\n", license.Confidence)
		fmt.Printf("  Compatible: %t\n", license.IsCompatible)
		if license.Link != "" {
			fmt.Printf("  Link: %s\n", license.Link)
		}
		fmt.Println()
	}

	slog.InfoContext(ctx, "License detection completed successfully", "licenses_count", len(licenses))
	return nil
}
