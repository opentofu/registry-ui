package main

import (
	"context"
	"fmt"
	"log"
	"log/slog"
	"os"
	"os/signal"
	"syscall"

	"github.com/urfave/cli/v3"

	dltofunightly "github.com/opentofu/registry-ui/command/dl-tofu-nightly"
	getmodulelicense "github.com/opentofu/registry-ui/command/get-module-license"
	getproviderlicense "github.com/opentofu/registry-ui/command/get-provider-license"
	rebuildglobalindexes "github.com/opentofu/registry-ui/command/rebuild-global-indexes"
	removeproviderversion "github.com/opentofu/registry-ui/command/remove-provider-version"
	retryversion "github.com/opentofu/registry-ui/command/retry-version"
	skipversion "github.com/opentofu/registry-ui/command/skip-version"
	syncmodule "github.com/opentofu/registry-ui/command/sync-module"
	syncmodules "github.com/opentofu/registry-ui/command/sync-modules"
	syncprovider "github.com/opentofu/registry-ui/command/sync-provider"
	syncproviders "github.com/opentofu/registry-ui/command/sync-providers"
	syncrepostats "github.com/opentofu/registry-ui/command/sync-repo-stats"
	"github.com/opentofu/registry-ui/pkg/config"
	"github.com/opentofu/registry-ui/pkg/db"
	"github.com/opentofu/registry-ui/pkg/telemetry"
)

// configFreeCommands lists subcommands that don't require config, telemetry, or DB.
var configFreeCommands = map[string]bool{
	"dl-tofu-nightly": true,
}

func main() {
	ctx := context.Background()

	// Setup structured logging first
	telemetry.SetupLogger()

	shutdown := func() {}

	app := &cli.Command{
		Name:     "registry-ui",
		Usage:    "OpenTofu Registry Backend CLI",
		Metadata: map[string]interface{}{},
		Before: func(ctx context.Context, cmd *cli.Command) (context.Context, error) {
			// Skip config/telemetry setup for commands that don't need it
			if cmd.Args().Present() && configFreeCommands[cmd.Args().First()] {
				return ctx, nil
			}

			backendConfig, err := config.LoadConfig(ctx)
			if err != nil {
				return ctx, fmt.Errorf("error loading config: %w", err)
			}
			config.StoreToCLI(cmd, backendConfig)

			// Initialize telemetry with config
			ctx, s, err := telemetry.SetupTelemetry(ctx, backendConfig.Telemetry)
			if err != nil {
				return ctx, fmt.Errorf("failed to initialize telemetry: %w", err)
			}
			shutdown = s

			if backendConfig.Telemetry.Enabled {
				// Ensure that our OTLP exporters can connect correctly
				if err := telemetry.TestOTLPConnection(ctx, backendConfig.Telemetry); err != nil {
					return ctx, fmt.Errorf("error testing OTLP connection: %w", err)
				}
			}

			// Setup signal handling for graceful shutdown
			c := make(chan os.Signal, 1)
			signal.Notify(c, os.Interrupt, syscall.SIGTERM)
			go func() {
				<-c
				slog.InfoContext(ctx, "Received interrupt signal, shutting down gracefully...")
				shutdown()
				os.Exit(0)
			}()

			return ctx, nil
		},
		After: func(ctx context.Context, cmd *cli.Command) error {
			shutdown()
			return nil
		},
		Commands: []*cli.Command{
			syncprovider.NewCommand(),
			syncproviders.NewCommand(),
			syncmodule.NewCommand(),
			syncmodules.NewCommand(),
			syncrepostats.NewCommand(),
			getmodulelicense.NewCommand(),
			getproviderlicense.NewCommand(),
			rebuildglobalindexes.NewCommand(),
			skipversion.NewCommand(),
			retryversion.NewCommand(),
			removeproviderversion.NewCommand(),
			db.NewMigrateCommand(),
			dltofunightly.NewCommand(),
		},
	}

	err := app.Run(ctx, os.Args)
	if err != nil {
		slog.ErrorContext(ctx, "Error running app", "error", err)
		log.Fatalf("Error: %v", err)
	}
}
