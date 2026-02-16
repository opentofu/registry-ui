package main

import (
	"context"
	"log"
	"log/slog"
	"os"
	"os/signal"
	"syscall"

	"github.com/urfave/cli/v3"

	dltofunightly "github.com/opentofu/registry-ui/cmd/dl-tofu-nightly"
	getmodulelicense "github.com/opentofu/registry-ui/cmd/get-module-license"
	getproviderlicense "github.com/opentofu/registry-ui/cmd/get-provider-license"
	rebuildglobalindexes "github.com/opentofu/registry-ui/cmd/rebuild-global-indexes"
	removeproviderversion "github.com/opentofu/registry-ui/cmd/remove-provider-version"
	retryversion "github.com/opentofu/registry-ui/cmd/retry-version"
	skipversion "github.com/opentofu/registry-ui/cmd/skip-version"
	syncmodule "github.com/opentofu/registry-ui/cmd/sync-module"
	syncmodules "github.com/opentofu/registry-ui/cmd/sync-modules"
	syncprovider "github.com/opentofu/registry-ui/cmd/sync-provider"
	syncproviders "github.com/opentofu/registry-ui/cmd/sync-providers"
	syncrepostats "github.com/opentofu/registry-ui/cmd/sync-repo-stats"
	"github.com/opentofu/registry-ui/pkg/config"
	"github.com/opentofu/registry-ui/pkg/db"
	"github.com/opentofu/registry-ui/pkg/telemetry"
)

func main() {
	ctx := context.Background()

	// Setup structured logging first
	telemetry.SetupLogger()

	backendConfig, err := config.LoadConfig(ctx)
	if err != nil {
		slog.ErrorContext(ctx, "Error loading config", "error", err)
		log.Fatalf("Error loading config: %v", err)
	}

	// Initialize telemetry with config
	ctx, shutdown, err := telemetry.SetupTelemetry(ctx, backendConfig.Telemetry)
	if err != nil {
		log.Fatalf("Failed to initialize telemetry: %v", err)
	}

	// Ensure that our logger can connect correctly
	err = telemetry.TestOTLPConnection(ctx)
	if err != nil {
		slog.ErrorContext(ctx, "Error testing OTLP connection", "error", err)
		log.Fatalf("Error testing OTLP connection: %v", err)
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

	defer shutdown()

	// Create CLI app with commands
	app := &cli.Command{
		Name:  "registry-ui",
		Usage: "OpenTofu Registry Backend CLI",
		Commands: []*cli.Command{
			syncprovider.NewCommand(backendConfig),
			syncproviders.NewCommand(backendConfig),
			syncmodule.NewCommand(backendConfig),
			syncmodules.NewCommand(backendConfig),
			syncrepostats.NewCommand(backendConfig),
			getmodulelicense.NewCommand(backendConfig),
			getproviderlicense.NewCommand(backendConfig),
			rebuildglobalindexes.NewCommand(backendConfig),
			skipversion.NewCommand(backendConfig),
			retryversion.NewCommand(backendConfig),
			removeproviderversion.NewCommand(backendConfig),
			db.NewMigrateCommand(backendConfig.DB),
			dltofunightly.NewCommand(backendConfig),
		},
	}

	err = app.Run(ctx, os.Args)
	if err != nil {
		slog.ErrorContext(ctx, "Error", "error", err)
		log.Fatalf("Error running app: %v", err)
	}
}
