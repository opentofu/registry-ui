package factory

import (
	"context"
	"crypto/tls"
	"crypto/x509"
	"fmt"
	"os"
	"path"
	"runtime"
	"strings"

	"github.com/ProtonMail/gopenpgp/v2/crypto"
	"github.com/opentofu/libregistry/logger"
	"github.com/opentofu/libregistry/metadata"
	"github.com/opentofu/libregistry/metadata/storage/filesystem"
	"github.com/opentofu/libregistry/vcs"
	"github.com/opentofu/libregistry/vcs/github"
	"github.com/opentofu/registry-ui/internal"
	"github.com/opentofu/registry-ui/internal/indexstorage"
	"github.com/opentofu/registry-ui/internal/indexstorage/bufferedstorage"
	"github.com/opentofu/registry-ui/internal/indexstorage/filesystemstorage"
	"github.com/opentofu/registry-ui/internal/indexstorage/s3storage"
	"github.com/opentofu/registry-ui/internal/license"
	"github.com/opentofu/registry-ui/internal/moduleindex"
	"github.com/opentofu/registry-ui/internal/moduleindex/moduleschema"
	"github.com/opentofu/registry-ui/internal/providerindex"
	"github.com/opentofu/registry-ui/internal/providerindex/providerdocsource"
	"github.com/opentofu/registry-ui/internal/providerindex/providerindexstorage"
	"github.com/opentofu/registry-ui/internal/registrycloner"
	"github.com/opentofu/registry-ui/internal/search"
	"github.com/opentofu/registry-ui/internal/search/searchstorage/indexstoragesearch"
	"github.com/opentofu/registry-ui/internal/server"
	"github.com/opentofu/tofudl"
	"github.com/opentofu/tofudl/branding"
)

type S3Parameters struct {
	CACertFile string
	Bucket     string
	Endpoint   string
	PathStyle  bool
	AccessKey  string
	SecretKey  string
	Region     string
}

func New(log logger.Logger) (BackendFactory, error) {
	return &backendFactory{
		logger: log,
	}, nil
}

type BackendFactory interface {
	Create(ctx context.Context, registryDir string, workDir string, destinationDir string, s3Params S3Parameters, parallelism int, tofuBinaryPath string) (internal.Backend, error)
}

type backendFactory struct {
	logger logger.Logger
}

func (b backendFactory) Create(ctx context.Context, registryDir string, workDir string, destinationDir string, s3Params S3Parameters, parallelism int, tofuBinaryPath string) (internal.Backend, error) {
	return getBackend(ctx, b.logger, registryDir, workDir, destinationDir, s3Params, parallelism, tofuBinaryPath)
}

func getBackend(ctx context.Context, log logger.Logger, registryDir string, workDir string, destinationDir string, s3Params S3Parameters, parallelism int, tofuBinaryPath string) (internal.Backend, error) {
	reader, err := metadata.New(filesystem.New(registryDir))
	if err != nil {
		return nil, err
	}

	licenseDetector, err := license.New(
		license.WithOSIApprovedLicenses(),
		license.WithCompatibleLicenses("MPL-2.0-no-copyleft-exception"),
		license.WithCompatibleLicenses("MIT-feh"),
	)
	if err != nil {
		return nil, err
	}

	vcsClient, err := github.New(
		// Skip cleaning up to allow for long-term reuse
		github.WithSkipCleanupWorkingCopyOnClose(true),
		github.WithCheckoutRootDirectory(workDir),
		github.WithLogger(log),
	)
	if err != nil {
		return nil, err
	}

	var backingStorage indexstorage.API
	if s3Params.Bucket != "" {
		var tlsPool *x509.CertPool
		if s3Params.CACertFile != "" {
			cacert, err := os.ReadFile(s3Params.CACertFile)
			if err != nil {
				return nil, fmt.Errorf("failed to read %s (%w)", s3Params.CACertFile, err)
			}
			tlsPool = x509.NewCertPool()
			if !tlsPool.AppendCertsFromPEM(cacert) {
				return nil, fmt.Errorf("failed to parse %s", s3Params.CACertFile)
			}
		} else {
			tlsPool, err = x509.SystemCertPool()
			if err != nil {
				return nil, fmt.Errorf("system cert pool not available (%w)", err)
			}
		}
		backingStorage, err = s3storage.New(
			ctx,
			s3storage.WithBucket(s3Params.Bucket),
			//s3storage.WithRegion(s3Params.Region),
			s3storage.WithEndpoint(s3Params.Endpoint),
			s3storage.WithTLSConfig(&tls.Config{
				RootCAs:    tlsPool,
				MinVersion: tls.VersionTLS12,
			}),
			s3storage.WithSecretKey(s3Params.SecretKey),
			s3storage.WithAccessKey(s3Params.AccessKey),
			s3storage.WithPathStyle(s3Params.PathStyle),
			s3storage.WithLogger(log),
		)
	} else {
		backingStorage, err = filesystemstorage.New(destinationDir)
	}
	if err != nil {
		return nil, err
	}

	// The buffered storage writes to a temporary directory and then tries to commit all changes to the docs dir or S3.
	// If the commit fails, this directory should be preserved for a future run.
	// TODO make this directory configurable.
	bufferedStorage, err := bufferedstorage.New(log, path.Join(workDir, ".docs-cache"), backingStorage, parallelism)
	if err != nil {
		return nil, err
	}

	searchStorage, err := indexstoragesearch.New(bufferedStorage)
	if err != nil {
		return nil, err
	}

	searchAPI, err := search.New(searchStorage, search.WithLogger(log))
	if err != nil {
		return nil, err
	}

	providerIndexGenerator, err := getProviderIndexer(ctx, bufferedStorage, log, reader, vcsClient, licenseDetector, searchAPI)
	if err != nil {
		return nil, err
	}

	moduleIndexGenerator, err := getModuleIndexGenerator(ctx, workDir, bufferedStorage, log, reader, vcsClient, licenseDetector, searchAPI, tofuBinaryPath)
	if err != nil {
		return nil, err
	}

	cloner, err := registrycloner.New(
		registrycloner.WithDirectory(registryDir),
	)
	if err != nil {
		return nil, err
	}

	openAPIWriter, err := server.NewWriter(bufferedStorage)
	if err != nil {
		return nil, err
	}

	return internal.New(cloner, moduleIndexGenerator, providerIndexGenerator, searchAPI, openAPIWriter, bufferedStorage, internal.WithLogger(log))
}

func getProviderIndexer(ctx context.Context, rootStorage indexstorage.API, log logger.Logger, reader metadata.API, vcsClient vcs.Client, licenseDetector license.Detector, searchAPI search.API) (providerindex.DocumentationGenerator, error) {
	storage, err := rootStorage.Subdirectory(ctx, "providers")
	if err != nil {
		return nil, err
	}
	destination, err := providerindexstorage.New(storage)
	if err != nil {
		return nil, err
	}
	source, err := providerdocsource.New(licenseDetector, log)
	if err != nil {
		return nil, err
	}
	scrapingManager := providerindex.NewDocumentationGenerator(log, reader, vcsClient, licenseDetector, source, destination, searchAPI)
	return scrapingManager, nil
}

func getModuleIndexGenerator(ctx context.Context, workDir string, rootStorage indexstorage.API, log logger.Logger, reader metadata.API, vcsClient vcs.Client, licenseDetector license.Detector, searchAPI search.API, tofuBinaryPath string) (moduleindex.Generator, error) {
	moduleStorage, err := rootStorage.Subdirectory(ctx, "modules")
	if err != nil {
		return nil, err
	}

	downloader, err := setupTofuDL(ctx, tofuBinaryPath)
	if err != nil {
		return nil, err
	}
	tofuBinary := path.Join(workDir, "tofu")
	if runtime.GOOS == "windows" {
		tofuBinary += ".exe"
	}
	moduleSchemaExtractor, err := moduleschema.NewExternalTofuExtractor(
		moduleschema.ExternalTofuExtractorConfig{
			TofuPath: tofuBinary,
		},
		log,
		downloader,
	)
	if err != nil {
		return nil, err
	}
	return moduleindex.New(
		log,
		reader,
		vcsClient,
		licenseDetector,
		moduleStorage,
		moduleSchemaExtractor,
		searchAPI,
	), nil
}

func setupTofuDL(ctx context.Context, tofuBinaryPath string) (tofudl.Mirror, error) {
	// Temporary hack until the patch gets merged into OpenTofu:
	key, err := crypto.GenerateKey(branding.ProductName+" Test", "noreply@example.org", "rsa", 2048)
	if err != nil {
		return nil, err
	}
	pubKey, err := key.GetArmoredPublicKey()
	if err != nil {
		return nil, err
	}
	builder, err := tofudl.NewReleaseBuilder(key)
	if err != nil {
		return nil, err
	}
	tofuBinaryContents, err := os.ReadFile(tofuBinaryPath)
	if err != nil {
		return nil, fmt.Errorf("failed to load tofu binary (did you run go generate? %w)", err)
	}
	if err := builder.PackageBinary(tofudl.PlatformAuto, tofudl.ArchitectureAuto, tofuBinaryContents, nil); err != nil {
		return nil, fmt.Errorf("failed to package tofu binary (%w)", err)
	}

	storagePath := path.Join(os.TempDir(), "tofu")
	if err := os.RemoveAll(storagePath); err != nil {
		return nil, fmt.Errorf("failed to clean storage path %s (%w)", storagePath, err)
	}
	if err := os.MkdirAll(storagePath, 0700); err != nil {
		return nil, fmt.Errorf("failed to create storage path %s (%w)", storagePath, err)
	}
	mirrorStorage, err := tofudl.NewFilesystemStorage(storagePath)
	if err != nil {
		return nil, fmt.Errorf("failed to set up tofudl mirror")
	}
	downloader, err := tofudl.NewMirror(
		tofudl.MirrorConfig{
			GPGKey: pubKey,
		},
		mirrorStorage,
		nil,
	)
	if err != nil {
		return nil, err
	}
	if err := builder.Build(ctx, "1.9.0", downloader); err != nil {
		// TODO: this is not cool.
		if !strings.Contains(err.Error(), "already exists") {
			return nil, fmt.Errorf("failed to push new tofu version (%w)", err)
		}
	}

	if _, err := downloader.Download(ctx); err != nil {
		return nil, fmt.Errorf("download failed (%w)", err)
	}

	return downloader, nil
}
