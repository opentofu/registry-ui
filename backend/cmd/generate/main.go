package main

import (
	"context"
	"encoding/json"
	"flag"
	"fmt"
	"log/slog"
	"os"
	"path"
	"regexp"
	"runtime"
	"strings"

	"github.com/opentofu/libregistry/logger"
	"github.com/opentofu/libregistry/types/module"
	backend "github.com/opentofu/registry-ui/internal"
	"github.com/opentofu/registry-ui/internal/blocklist"
	"github.com/opentofu/registry-ui/internal/defaults"
	"github.com/opentofu/registry-ui/internal/factory"
)

func main() {
	licensesFile := ""
	skipUpdateProviders := false
	skipUpdateModules := false
	namespacePrefix := ""
	namespace := ""
	name := ""
	targetSystem := ""
	logLevel := "info"
	forceRegenerate := ""
	registryDir := defaults.RegistryDir
	workDir := defaults.WorkDir
	destinationDir := defaults.DestinationDir
	commitParallelism := 100
	blockListFile := ""
	// TODO this is only until these features get released in mainline tofu.
	binaryName := "tofu"
	if runtime.GOOS == "windows" {
		binaryName += ".exe"
	}
	ghToken := os.Getenv("GITHUB_TOKEN")
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

	flag.StringVar(&licensesFile, "licenses-file", licensesFile, "JSON file containing a list of approved licenses to include when indexing. (required)")
	flag.BoolVar(&skipUpdateProviders, "skip-update-providers", skipUpdateProviders, "Skip updating provider indexes.")
	flag.BoolVar(&skipUpdateModules, "skip-update-modules", skipUpdateModules, "Skip updating module indexes.")
	flag.StringVar(&namespacePrefix, "namespace-prefix", namespace, "Limit updates to a namespace prefix.")
	flag.StringVar(&namespace, "namespace", namespace, "Limit updates to a namespace.")
	flag.StringVar(&name, "name", name, "Limit updates to a name. Only works in conjunction with -namespace. For providers, this will result in a single provider getting updated. For modules, this will update all target systems under a name.")
	flag.StringVar(&targetSystem, "target-system", targetSystem, "Limit updates to a target system for module updates only. Only works in conjunction with -namespace and -name.")
	flag.StringVar(&logLevel, "log-level", logLevel, "Set the log level (debug, info, warn, error)")
	flag.StringVar(&registryDir, "registry-dir", registryDir, "Directory to check out the registry in.")
	flag.StringVar(&workDir, "vcs-dir", workDir, "Directory to use for checking out providers and modules in.")
	flag.StringVar(&destinationDir, "destination-dir", destinationDir, "Directory to place generated documentation and indexes in.")
	flag.StringVar(&s3Params.Bucket, "s3-bucket", s3Params.Bucket, "S3 bucket to use for uploads.")
	flag.StringVar(&s3Params.Endpoint, "s3-endpoint", s3Params.Endpoint, "S3 endpoint to use for uploads. Defaults to AWS.")
	flag.BoolVar(&s3Params.PathStyle, "s3-path-style", s3Params.PathStyle, "Use path-style URLs for S3.")
	flag.StringVar(&s3Params.CACertFile, "s3-ca-cert-file", s3Params.CACertFile, "File containing the CA certificate for the S3 endpoint. Defaults to the system certificates.")
	flag.StringVar(&s3Params.Region, "s3-region", s3Params.Region, "Region to use for S3 uploads.")
	flag.IntVar(&commitParallelism, "commit-parallelism", commitParallelism, "Parallel uploads to use on commit.")
	flag.StringVar(&tofuBinaryPath, "tofu-binary-path", tofuBinaryPath, "Temporary: Tofu binary path to use for module schema extraction.")
	flag.StringVar(&forceRegenerate, "force-regenerate", forceRegenerate, "Force regenerating a namespace, name, or target system. This parameter is a comma-separate list consisting of either a namespace, a namespace and a name separated by a /, or a namespace, name and target system separated by a /. Example: namespace/name/targetsystem,othernamespace/othername")
	flag.StringVar(&blockListFile, "blocklist", blockListFile, "File containing the blocklist to use.")
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

	approvedLicenses, err := readLicensesFile(ctx, licensesFile)
	if err != nil {
		mainLogger.Error(ctx, err.Error())
		os.Exit(1)
	}

	mainLogger.Info(ctx, "Initializing metadata system...")

	blockList := blocklist.New()
	if blockListFile != "" {
		if err := blockList.LoadFile(blockListFile); err != nil {
			mainLogger.Error(ctx, "Failed to load block file %s (%v)", blockListFile, err)
			os.Exit(1)
		}
	}

	backendFactory, err := factory.New(mainLogger, ghToken)
	if err != nil {
		mainLogger.Error(ctx, err.Error())
		os.Exit(1)
	}

	backendInstance, err := backendFactory.Create(ctx, registryDir, workDir, destinationDir, blockList, s3Params, commitParallelism, tofuBinaryPath, approvedLicenses)
	if err != nil {
		mainLogger.Error(ctx, err.Error())
		os.Exit(1)
	}

	// TODO implement signal handling

	forceOpts, err := parseForceOpts(forceRegenerate)
	if err != nil {
		mainLogger.Error(ctx, err.Error())
		os.Exit(1)
	}

	if err := backendInstance.Generate(
		ctx,
		append(forceOpts,
			backend.WithNamespacePrefix(namespacePrefix),
			backend.WithNamespace(namespace),
			backend.WithName(name),
			backend.WithTargetSystem(targetSystem),
			backend.WithSkipUpdateModules(skipUpdateModules),
			backend.WithSkipUpdateProviders(skipUpdateProviders),
		)...,
	); err != nil {
		mainLogger.Error(ctx, err.Error())
		os.Exit(1)
	}

	mainLogger.Info(ctx, "Done!")
}

func readLicensesFile(_ context.Context, licensesFile string) ([]string, error) {
	if licensesFile == "" {
		return nil, fmt.Errorf("the --licenses-file parameter is required")
	}

	licensesFileContent, err := os.ReadFile(licensesFile)
	if err != nil {
		return nil, fmt.Errorf("failed to read licenses file: %w", err)
	}

	re := regexp.MustCompile(`//.*?\n`)
	licensesFileContent = re.ReplaceAll(licensesFileContent, nil)

	var approvedLicenses []string
	if err := json.Unmarshal(licensesFileContent, &approvedLicenses); err != nil {
		return nil, fmt.Errorf("failed to unmarshal licenses file: %w", err)
	}
	if len(approvedLicenses) == 0 {
		return nil, fmt.Errorf("the licenses file contains no licenses")
	}
	return approvedLicenses, nil
}

func parseForceOpts(regenerate string) ([]backend.GenerateOpt, error) {
	if regenerate == "" {
		return nil, nil
	}
	var results []backend.GenerateOpt
	parts := strings.Split(regenerate, ",")
	for _, part := range parts {
		if part == "" {
			return nil, fmt.Errorf("empty part in -force-regenerate options")
		}
		partParts := strings.Split(part, "/")
		switch len(partParts) {
		case 1:
			results = append(results, backend.WithForceRegenerateNamespace(partParts[0]))
		case 2:
			results = append(results, backend.WithForceRegenerateNamespaceAndName(partParts[0], partParts[1]))
		case 3:
			results = append(results, backend.WithForceRegenerateSingleModule(module.Addr{
				Namespace:    partParts[0],
				Name:         partParts[1],
				TargetSystem: partParts[2],
			}))
		default:
			return nil, fmt.Errorf("invalid option for -force-regenerate: %s", part)
		}
	}
	return results, nil
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
