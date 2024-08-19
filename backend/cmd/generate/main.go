package main

import (
	"context"
	"flag"
	"fmt"
	"log/slog"
	"os"
	"path"
	"runtime"

	"github.com/opentofu/libregistry/logger"
	backend "github.com/opentofu/registry-ui/internal"
	"github.com/opentofu/registry-ui/internal/defaults"
	"github.com/opentofu/registry-ui/internal/factory"
)

func main() {
	skipUpdateProviders := false
	skipUpdateModules := false
	namespace := ""
	logLevel := "info"
	registryDir := defaults.RegistryDir
	workDir := defaults.WorkDir
	destinationDir := defaults.DestinationDir
	commitParallelism := 25
	// TODO this is only until these features get released in mainline tofu.
	binaryName := "tofu"
	if runtime.GOOS == "windows" {
		binaryName += ".exe"
	}
	tofuBinaryPath := path.Join("internal", "moduleindex", "moduleschema", "testtofu", binaryName)
	s3Params := factory.S3Parameters{
		AccessKey:  os.Getenv("AWS_ACCESS_KEY_ID"),
		SecretKey:  os.Getenv("AWS_SECRET_ACCESS_KEY"),
		Region:     os.Getenv("AWS_REGION"),
		CACertFile: os.Getenv("AWS_CA_BUNDLE"),
		Endpoint:   os.Getenv("AWS_ENDPOINT_URL_S3"),
	}
	if s3Params.Endpoint == "" {
		s3Params.Endpoint = os.Getenv("AWS_ENDPOINT_URL")
	}

	flag.BoolVar(&skipUpdateProviders, "skip-update-providers", skipUpdateProviders, "Skip updating provider indexes.")
	flag.BoolVar(&skipUpdateModules, "skip-update-modules", skipUpdateModules, "Skip updating module indexes.")
	flag.StringVar(&namespace, "namespace", namespace, "Limit updates to a namespace.")
	flag.StringVar(&logLevel, "log-level", logLevel, "Set the log level (debug, info, warn, error)")
	flag.StringVar(&registryDir, "registry-dir", registryDir, "Directory to check out the registry in.")
	flag.StringVar(&workDir, "vcs-dir", workDir, "Directory to use for checking out providers and modules in.")
	flag.StringVar(&destinationDir, "destination-dir", destinationDir, "Directory to place generated documentation and indexes in.")
	flag.StringVar(&s3Params.Bucket, "s3-bucket", s3Params.Bucket, "S3 bucket to use for uploads.")
	flag.StringVar(&s3Params.Endpoint, "s3-endpoint", s3Params.Endpoint, "S3 endpoint to use for uploads. Defaults to AWS.")
	flag.BoolVar(&s3Params.PathStyle, "s3-path-style", s3Params.PathStyle, "Use path-style URLs for S3.")
	flag.StringVar(&s3Params.CACertFile, "s3-ca-cert-file", s3Params.CACertFile, "File containing the CA certificate for the S3 endpoint. Defaults to the system certificates.")
	flag.StringVar(&s3Params.Region, "s3-region", s3Params.Region, "Region to use for S3 uploads.")
	flag.IntVar(&commitParallelism, "parallelism", commitParallelism, "Parallel uploads to use on commit.")
	flag.StringVar(&tofuBinaryPath, "tofu-binary-path", tofuBinaryPath, "Temporary: Tofu binary path to use for module schema extraction.")
	flag.Parse()

	// Parse the log level
	level, err := parseLogLevel(logLevel)
	if err != nil {
		_, _ = os.Stderr.Write([]byte(err.Error()))
		os.Exit(1)
	}
	mainLogger := logger.NewSLogLogger(slog.New(slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{
		Level: level,
	})))

	ctx := context.Background()

	mainLogger.Info(ctx, "Initializing metadata system...")

	backendFactory, err := factory.New(mainLogger)
	if err != nil {
		mainLogger.Error(ctx, err.Error())
		os.Exit(1)
	}

	backendInstance, err := backendFactory.Create(ctx, registryDir, workDir, destinationDir, s3Params, commitParallelism, tofuBinaryPath)
	if err != nil {
		mainLogger.Error(ctx, err.Error())
		os.Exit(1)
	}

	// TODO implement signal handling

	if err := backendInstance.Generate(
		ctx,
		backend.WithNamespace(namespace),
		backend.WithSkipUpdateModules(skipUpdateModules),
		backend.WithSkipUpdateProviders(skipUpdateProviders),
	); err != nil {
		mainLogger.Error(ctx, err.Error())
		os.Exit(1)
	}

	mainLogger.Info(ctx, "Done!")
}

func parseLogLevel(level string) (slog.Level, error) {
	switch level {
	case "trace":
		// This is a custom log level in libregistry.
		return -8, nil
	case "debug":
		return slog.LevelDebug, nil
	case "info":
		return slog.LevelInfo, nil
	case "warn":
		return slog.LevelWarn, nil
	case "error":
		return slog.LevelError, nil
	default:
		return slog.LevelDebug, fmt.Errorf("unknown log level: %s", level)
	}
}
