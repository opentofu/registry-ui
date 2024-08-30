package internal_test

import (
	"os"
	"path"
	"runtime"
	"testing"

	"github.com/opentofu/libregistry/logger"
	"github.com/opentofu/registry-ui/internal"
	"github.com/opentofu/registry-ui/internal/blocklist"
	"github.com/opentofu/registry-ui/internal/factory"
	"github.com/opentofu/tofutestutils"
)

func TestE2E(t *testing.T) {
	aws := tofutestutils.AWS(t)

	log := logger.NewTestLogger(t)
	ctx := tofutestutils.Context(t)

	testDir := t.TempDir()
	registryDir := path.Join(testDir, "registry")
	workDir := path.Join(testDir, "work")
	docsDir := path.Join(testDir, "docs")

	tofutestutils.Must(os.MkdirAll(registryDir, 0755))
	tofutestutils.Must(os.MkdirAll(workDir, 0755))
	tofutestutils.Must(os.MkdirAll(docsDir, 0755))

	s3Params := factory.S3Parameters{
		CACertFile: aws.CACertFile(),
		Bucket:     aws.S3Bucket(),
		Endpoint:   aws.S3Endpoint(),
		PathStyle:  aws.S3UsePathStyle(),
		AccessKey:  aws.AccessKey(),
		SecretKey:  aws.SecretKey(),
		Region:     aws.Region(),
	}

	backendFactory := tofutestutils.Must2(factory.New(log))
	binaryName := "tofu"
	if runtime.GOOS == "windows" {
		binaryName += ".exe"
	}
	// TODO this is only until these features get released in mainline tofu.
	tofuBinaryPath := path.Join("moduleindex", "moduleschema", "testtofu", binaryName)
	backendInstance := tofutestutils.Must2(backendFactory.Create(ctx, registryDir, workDir, docsDir, blocklist.New(), s3Params, 25, tofuBinaryPath))

	tofutestutils.Must(backendInstance.Generate(
		ctx,
		internal.WithNamespace("integrations"),
	))

	// TODO check a few files if they are present on the S3 backend.
}

func TestE2EDoubleRun(t *testing.T) {
	aws := tofutestutils.AWS(t)

	log := logger.NewTestLogger(t)
	ctx := tofutestutils.Context(t)

	testDir := t.TempDir()
	registryDir := path.Join(testDir, "registry")
	workDir := path.Join(testDir, "work")
	docsDir := path.Join(testDir, "docs")

	tofutestutils.Must(os.MkdirAll(registryDir, 0755))
	tofutestutils.Must(os.MkdirAll(workDir, 0755))
	tofutestutils.Must(os.MkdirAll(docsDir, 0755))

	s3Params := factory.S3Parameters{
		CACertFile: aws.CACertFile(),
		Bucket:     aws.S3Bucket(),
		Endpoint:   aws.S3Endpoint(),
		PathStyle:  aws.S3UsePathStyle(),
		AccessKey:  aws.AccessKey(),
		SecretKey:  aws.SecretKey(),
		Region:     aws.Region(),
	}

	backendFactory := tofutestutils.Must2(factory.New(log))
	binaryName := "tofu"
	if runtime.GOOS == "windows" {
		binaryName += ".exe"
	}
	// TODO this is only until these features get released in mainline tofu.
	tofuBinaryPath := path.Join("moduleindex", "moduleschema", "testtofu", binaryName)
	backendInstance := tofutestutils.Must2(backendFactory.Create(ctx, registryDir, workDir, docsDir, blocklist.New(), s3Params, 25, tofuBinaryPath))

	t.Logf("🏃 Starting first run...")
	tofutestutils.Must(backendInstance.Generate(
		ctx,
		internal.WithNamespace("integrations"),
	))

	// TODO check if the second run is actually using the files on the storage. Maybe via metrics collection?

	t.Logf("🏃 Starting second run...")
	tofutestutils.Must(backendInstance.Generate(
		ctx,
		internal.WithNamespace("integrations"),
	))

	// TODO check a few files if they are present on the S3 backend.
}
