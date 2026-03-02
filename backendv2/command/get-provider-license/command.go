// Package getproviderlicense implements the get-provider-license command to test license detection for a specific provider.
package getproviderlicense

import (
	"context"
	"fmt"
	"log/slog"

	"github.com/urfave/cli/v3"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/codes"

	"github.com/opentofu/registry-ui/pkg/config"
	"github.com/opentofu/registry-ui/pkg/provider"
	"github.com/opentofu/registry-ui/pkg/telemetry"
)

func NewCommand() *cli.Command {
	return &cli.Command{
		Name:  "get-provider-license",
		Usage: "Test license detection for a specific provider",
		Flags: []cli.Flag{
			&cli.StringFlag{
				Name:     "namespace",
				Usage:    "Provider namespace (e.g., hashicorp)",
				Required: true,
			},
			&cli.StringFlag{
				Name:     "name",
				Usage:    "Provider name (e.g., aws)",
				Required: true,
			},
			&cli.StringFlag{
				Name:     "version",
				Aliases:  []string{"v"},
				Usage:    "Provider version (e.g., 5.0.0)",
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
	ctx, span := telemetry.Tracer().Start(ctx, "cmd.get_provider_license")
	defer span.End()

	namespace := cmd.String("namespace")
	name := cmd.String("name")
	version := cmd.String("version")

	span.SetAttributes(
		attribute.String("provider.namespace", namespace),
		attribute.String("provider.name", name),
		attribute.String("provider.version", version),
	)

	slog.InfoContext(ctx, "License detection for provider",
		"namespace", namespace, "name", name, "version", version)

	// Create provider reader
	providerReader, err := provider.NewProviderReader(ctx, cfg)
	if err != nil {
		span.RecordError(err)
		span.SetStatus(codes.Error, err.Error())
		slog.ErrorContext(ctx, "Failed to create provider reader", "error", err)
		return fmt.Errorf("failed to create provider reader: %w", err)
	}

	// Checkout the version for processing
	workDir, cleanup, err := providerReader.CheckoutVersionForScraping(ctx, namespace, name, version)
	if err != nil {
		span.RecordError(err)
		span.SetStatus(codes.Error, err.Error())
		slog.ErrorContext(ctx, "Failed to checkout version", "error", err)
		return fmt.Errorf("failed to checkout version: %w", err)
	}
	defer cleanup()

	slog.InfoContext(ctx, "Checked out provider version", "workDir", workDir)

	// Detect licenses in the directory
	licenses, err := providerReader.DetectLicensesInDirectory(ctx, namespace, name, workDir)
	if err != nil {
		span.RecordError(err)
		span.SetStatus(codes.Error, err.Error())
		slog.ErrorContext(ctx, "Failed to detect licenses", "error", err)
		return fmt.Errorf("failed to detect licenses: %w", err)
	}

	// Display results
	fmt.Printf("\n=== License Detection Results ===\n")
	fmt.Printf("Provider: %s/%s@%s\n", namespace, name, version)
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
