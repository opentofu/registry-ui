package main

import (
	"bufio"
	"context"
	"crypto/tls"
	"crypto/x509"
	"flag"
	"fmt"
	"log/slog"
	"os"
	"strings"

	"github.com/opentofu/libregistry/logger"
	"github.com/opentofu/libregistry/types/provider"

	"github.com/opentofu/registry-ui/internal/defaults"
	"github.com/opentofu/registry-ui/internal/factory"
	"github.com/opentofu/registry-ui/internal/indexstorage"
	"github.com/opentofu/registry-ui/internal/indexstorage/bufferedstorage"
	"github.com/opentofu/registry-ui/internal/indexstorage/filesystemstorage"
	"github.com/opentofu/registry-ui/internal/indexstorage/s3storage"
	"github.com/opentofu/registry-ui/internal/providerindex/providerindexstorage"
	"github.com/opentofu/registry-ui/internal/search"
	"github.com/opentofu/registry-ui/internal/search/searchstorage/indexstoragesearch"
	"github.com/opentofu/registry-ui/internal/search/searchtypes"
)

func main() {
	var (
		resourceType   string
		resourceAddr   string
		version        string
		force          bool
		dryRun         bool
		logLevel       = "info"
		destinationDir = defaults.DestinationDir
	)
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

	flag.StringVar(&logLevel, "log-level", logLevel, "Set the log level (debug, info, warn, error)")
	flag.StringVar(&destinationDir, "destination-dir", destinationDir, "Directory to place generated documentation and indexes in.")
	flag.StringVar(&s3Params.Bucket, "s3-bucket", s3Params.Bucket, "S3 bucket to use for uploads.")
	flag.StringVar(&s3Params.Endpoint, "s3-endpoint", s3Params.Endpoint, "S3 endpoint to use for uploads. Defaults to AWS.")
	flag.BoolVar(&s3Params.PathStyle, "s3-path-style", s3Params.PathStyle, "Use path-style URLs for S3.")
	flag.StringVar(&s3Params.CACertFile, "s3-ca-cert-file", s3Params.CACertFile, "File containing the CA certificate for the S3 endpoint. Defaults to the system certificates.")
	flag.StringVar(&s3Params.Region, "s3-region", s3Params.Region, "Region to use for S3 uploads.")
	flag.StringVar(&version, "version", version, "Specific version to remove (optional, removes all versions if not specified)")
	flag.BoolVar(&force, "force", force, "Skip confirmation prompt")
	flag.BoolVar(&dryRun, "dry-run", dryRun, "Show what would be removed without actually removing")
	flag.Parse()

	// Parse command arguments
	args := flag.Args()
	if len(args) < 2 {
		fmt.Fprintf(os.Stderr, "Usage: %s <resource-type> <resource-address> [flags]\n", os.Args[0])
		fmt.Fprintf(os.Stderr, "Resource types: provider\n")
		fmt.Fprintf(os.Stderr, "Example: %s provider hashicorp/aws\n", os.Args[0])
		os.Exit(1)
	}

	resourceType = args[0]
	resourceAddr = args[1]

	// Validate resource type
	if resourceType != "provider" {
		fmt.Fprintf(os.Stderr, "Error: unsupported resource type '%s'. Only 'provider' is currently supported.\n", resourceType)
		os.Exit(1)
	}

	// Parse the log level
	level, err := parseLogLevel(logLevel)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error: %s\n", err.Error())
		os.Exit(1)
	}
	mainLogger := logger.NewSLogLogger(slog.New(slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{
		Level: level,
	})))

	ctx := context.Background()

	// TODO: Extend to modules too, not just providers

	// Parse and validate provider address
	providerAddr := provider.Addr{
		Namespace: strings.Split(resourceAddr, "/")[0],
		Name:      strings.Split(resourceAddr, "/")[1],
	}
	if len(strings.Split(resourceAddr, "/")) != 2 {
		mainLogger.Error(ctx, "Invalid provider address format. Expected: namespace/name")
		os.Exit(1)
	}

	mainLogger.Info(ctx, "Initializing storage systems...")

	// Create temporary directory for buffered storage
	localDir, err := os.MkdirTemp("", "registry-remove-*")
	if err != nil {
		mainLogger.Error(ctx, "Failed to create temporary directory: %v", err)
		os.Exit(1)
	}
	defer func() {
		if err := os.RemoveAll(localDir); err != nil {
			mainLogger.Warn(ctx, "Failed to clean up temporary directory %s: %v", localDir, err)
		}
	}()

	backingStorage, err := getBackingStorage(ctx, mainLogger, s3Params, destinationDir)
	if err != nil {
		os.Exit(1)
	}

	bufferedStorage, err := bufferedstorage.New(mainLogger, localDir, backingStorage, 100)
	if err != nil {
		mainLogger.Error(ctx, "Failed to create buffered storage: %v", err)
		os.Exit(1)
	}

	providerStorageDir, err := bufferedStorage.Subdirectory(ctx, "providers")
	if err != nil {
		mainLogger.Error(ctx, "Failed to get provider storage directory: %v", err)
		os.Exit(1)
	}
	providerStorage, err := providerindexstorage.New(providerStorageDir)
	if err != nil {
		mainLogger.Error(ctx, "Failed to create provider storage: %v", err)
		os.Exit(1)
	}

	searchStorage, err := indexstoragesearch.New(bufferedStorage)
	if err != nil {
		mainLogger.Error(ctx, "Failed to create search storage: %v", err)
		os.Exit(1)
	}

	searchAPI, err := search.New(searchStorage, search.WithLogger(mainLogger))
	if err != nil {
		mainLogger.Error(ctx, "Failed to create search API: %v", err)
		os.Exit(1)
	}

	// Check if provider exists and get its versions
	providerData, err := providerStorage.GetProvider(ctx, providerAddr)
	if err != nil {
		mainLogger.Error(ctx, "Provider %s not found in registry: %v", providerAddr, err)
		// if the provider does not exist, it may still exist in the search index, so we try to remove it from there too
		if err := searchAPI.RemoveItem(ctx, searchtypes.IndexID("providers/"+providerAddr.String())); err != nil {
			mainLogger.Error(ctx, "Failed to remove provider from search index: %v", err)
			os.Exit(0)
		}
		mainLogger.Info(ctx, "Provider %s does not exist in the registry, nothing to remove", providerAddr)
		os.Exit(0)
	}

	var removeMessage string
	if version != "" {
		removeMessage = fmt.Sprintf("provider %s version %s", providerAddr, version)
	} else {
		removeMessage = fmt.Sprintf("provider %s and ALL its versions", providerAddr)
	}

	// Dry run mode
	if dryRun {
		mainLogger.Info(ctx, "DRY RUN: Would remove %s", removeMessage)

		if version == "" {
			// List all versions that would be removed
			mainLogger.Info(ctx, "DRY RUN: Would remove %d versions:", len(providerData.Versions))
			for _, v := range providerData.Versions {
				mainLogger.Info(ctx, "  - %s", v.ID)
			}
		} else {
			mainLogger.Info(ctx, "DRY RUN: No changes made")
		}
		os.Exit(0)
	}

	// Confirmation prompt (unless --force)
	if !force {
		fmt.Printf("Are you sure you want to remove %s? [y/N]: ", removeMessage)
		reader := bufio.NewReader(os.Stdin)
		response, err := reader.ReadString('\n')
		if err != nil {
			mainLogger.Error(ctx, "Failed to read response: %v", err)
			os.Exit(1)
		}

		response = strings.TrimSpace(strings.ToLower(response))
		if response != "y" && response != "yes" {
			mainLogger.Info(ctx, "Operation cancelled")
			os.Exit(0)
		}
	}

	// Perform the removal
	var versionsToRemove []provider.VersionNumber
	if version != "" {
		versionsToRemove = append(versionsToRemove, provider.VersionNumber(version))
		mainLogger.Info(ctx, "Removing provider %s version %s...", providerAddr, version)
	} else {
		for _, v := range providerData.Versions {
			versionsToRemove = append(versionsToRemove, v.ID)
		}
		mainLogger.Info(ctx, "Removing provider %s and all %d versions...", providerAddr, len(versionsToRemove))
	}

	// Remove each version from search index
	for _, v := range versionsToRemove {
		mainLogger.Debug(ctx, "Removing version %s from search index...", v)
		if err := searchAPI.RemoveVersionItems(ctx, searchtypes.IndexTypeProvider, providerAddr.String(), string(v)); err != nil {
			mainLogger.Error(ctx, "Failed to remove version %s from search index: %v", v, err)
			os.Exit(1)
		}
		if err := providerStorage.DeleteProviderVersion(ctx, providerAddr, v); err != nil {
			mainLogger.Error(ctx, "Failed to remove provider version: %v", err)
			os.Exit(1)
		}
	}

	// Remove from provider storage when we dont specify a version
	if version == "" {
		if err := providerStorage.DeleteProvider(ctx, providerAddr); err != nil {
			mainLogger.Error(ctx, "Failed to remove provider: %v", err)
			os.Exit(1)
		}
		mainLogger.Info(ctx, "Successfully removed provider %s (%d versions) docs", providerAddr, len(versionsToRemove))

		if err := searchAPI.RemoveItem(ctx, searchtypes.IndexID("providers/"+providerAddr.String())); err != nil {
			mainLogger.Error(ctx, "Failed to remove provider from search index: %v", err)
			os.Exit(1)
		}
	}

	// Commit the changes
	mainLogger.Info(ctx, "Committing changes...")
	if err := bufferedStorage.Commit(ctx); err != nil {
		mainLogger.Error(ctx, "Failed to commit changes: %v", err)
		os.Exit(1)
	}

	mainLogger.Info(ctx, "Done!")
}

func getBackingStorage(ctx context.Context, mainLogger logger.Logger, s3Params factory.S3Parameters, destinationDir string) (indexstorage.API, error) {
	var backingStorage indexstorage.API
	var err error
	if s3Params.Bucket != "" {
		var tlsPool *x509.CertPool
		if s3Params.CACertFile != "" {
			cacert, err := os.ReadFile(s3Params.CACertFile)
			if err != nil {
				mainLogger.Error(ctx, "Failed to read CA cert file: %v", err)
				os.Exit(1)
			}
			tlsPool = x509.NewCertPool()
			if !tlsPool.AppendCertsFromPEM(cacert) {
				mainLogger.Error(ctx, "Failed to parse CA cert")
				os.Exit(1)
			}
		} else {
			tlsPool, err = x509.SystemCertPool()
			if err != nil {
				mainLogger.Error(ctx, "System cert pool not available: %v", err)
				os.Exit(1)
			}
		}
		backingStorage, err = s3storage.New(
			ctx,
			s3storage.WithBucket(s3Params.Bucket),
			s3storage.WithRegion(s3Params.Region),
			s3storage.WithEndpoint(s3Params.Endpoint),
			s3storage.WithPathStyle(s3Params.PathStyle),
			s3storage.WithTLSConfig(&tls.Config{
				RootCAs: tlsPool,
			}),
			s3storage.WithAccessKey(s3Params.AccessKey),
			s3storage.WithSecretKey(s3Params.SecretKey),
		)
		if err != nil {
			mainLogger.Error(ctx, "Failed to create S3 storage: %v", err)
			os.Exit(1)
		}
	} else {
		backingStorage, err = filesystemstorage.New(destinationDir)
		if err != nil {
			mainLogger.Error(ctx, "Failed to create filesystem storage: %v", err)
			os.Exit(1)
		}
	}
	return backingStorage, err
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
