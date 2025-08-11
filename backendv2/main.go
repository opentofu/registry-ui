package main

import (
	"context"
	"log"
	"log/slog"
	"os"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/s3"
	"github.com/urfave/cli/v3"
	"go.opentelemetry.io/otel"
	
	syncprovider "github.com/opentofu/registry-ui/cmd/sync-provider"
	syncproviders "github.com/opentofu/registry-ui/cmd/sync-providers"
	"github.com/opentofu/registry-ui/pkg/config"
	"github.com/opentofu/registry-ui/pkg/db"
)

func main() {
	ctx := context.Background()

	// Setup structured logging first
	setupLogger()

	backendConfig, err := config.LoadConfig()
	if err != nil {
		slog.ErrorContext(ctx, "Error loading config", "error", err)
		log.Fatalf("Error loading config: %v", err)
	}

	// Initialize telemetry with config
	ctx, shutdown, err := initTelemetry(ctx, backendConfig.Telemetry)
	if err != nil {
		log.Fatalf("Failed to initialize telemetry: %v", err)
	}
	defer shutdown()

	// Create tracer
	tracer := otel.Tracer("opentofu-registry-backend")
	ctx, span := tracer.Start(ctx, "main")
	defer span.End()

	// Create CLI app with commands
	app := &cli.Command{
		Name:  "registry-ui",
		Usage: "OpenTofu Registry Backend CLI",
		Commands: []*cli.Command{
			syncprovider.NewCommand(backendConfig),
			syncproviders.NewCommand(backendConfig),
			db.NewMigrateCommand(backendConfig.DB),
		},
		Before: func(ctx context.Context, cmd *cli.Command) (context.Context, error) {
			// Add config to context for subcommands
			return context.WithValue(ctx, "config", backendConfig), nil
		},
	}

	if err := testBucketConnection(ctx, backendConfig); err != nil {
		slog.ErrorContext(ctx, "Failed to connect to bucket", "error", err)
		log.Fatalf("Failed to connect to bucket: %v", err)
	}
	slog.InfoContext(ctx, "Successfully connected to bucket")

	if err := testDBConnection(ctx, &backendConfig.DB); err != nil {
		slog.ErrorContext(ctx, "Failed to connect to database", "error", err)
		log.Fatalf("Failed to connect to database: %v", err)
	}
	slog.InfoContext(ctx, "Successfully connected to database")

	app.Run(ctx, os.Args)
}

func testBucketConnection(ctx context.Context, backendConfig *config.BackendConfig) error {
	tracer := otel.Tracer("opentofu-registry-backend")
	ctx, span := tracer.Start(ctx, "testBucketConnection")
	defer span.End()

	slog.InfoContext(ctx, "Testing bucket connection",
		"bucket", backendConfig.Bucket.BucketName,
		"endpoint", backendConfig.Bucket.Endpoint, // TODO: is this sensitive info?
		"region", backendConfig.Bucket.Region)

	bucketCfg := &backendConfig.Bucket
	client, err := bucketCfg.GetClient(ctx)
	if err != nil {
		slog.ErrorContext(ctx, "Failed to connect to bucket", "error", err)
	}

	// Test connection by checking if bucket exists
	_, err = client.HeadBucket(ctx, &s3.HeadBucketInput{
		Bucket: aws.String(backendConfig.Bucket.BucketName),
	})

	if err != nil {
		slog.ErrorContext(ctx, "Bucket connection test failed", "error", err)
	} else {
		slog.InfoContext(ctx, "Bucket connection test successful")
	}

	return err
}

func testDBConnection(ctx context.Context, dbConfig *config.DBConfig) error {
	tracer := otel.Tracer("opentofu-registry-backend")
	ctx, span := tracer.Start(ctx, "testDBConnection")
	defer span.End()

	slog.InfoContext(ctx, "Testing database connection")

	pool, err := dbConfig.GetPool(ctx)
	if err != nil {
		slog.ErrorContext(ctx, "Failed to connect to database", "error", err)
	} else {
		// Test the connection with a ping
		err = pool.Ping(ctx)
		if err != nil {
			slog.ErrorContext(ctx, "Database ping failed", "error", err)
		} else {
			slog.InfoContext(ctx, "Database connection test successful")
		}
	}

	return err
}
