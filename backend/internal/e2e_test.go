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
	"github.com/opentofu/registry-ui/internal/testutil"
	"github.com/opentofu/tofutestutils"
)

var testLicenseList = []string{
	"AFL-1.1", "AFL-1.2", "AFL-2.0", "AFL-2.1", "AFL-3.0",
	"Apache-1.1", "Apache-2.0",
	"Artistic-1.0", "Artistic-1.0-Perl", "Artistic-1.0-cl8", "Artistic-2.0",
	"0BSD", "BSD-1-Clause", "BSD-2-Clause", "BSD-2-Clause-Patent", "BSD-3-Clause", "BSD-3-Clause-LBNL",
	"CDDL-1.0",
	"EPL-1.0", "EPL-2.0",
	"ICU",
	"ISC",
	"MIT", "MIT-0", "MIT-Modern-Variant", "MIT-feh",
	"MPL-1.0", "MPL-1.1", "MPL-2.0", "MPL-2.0-no-copyleft-exception",
	"Unlicense",
	"Xnet",
	"Zlib",
}

func TestE2E(t *testing.T) {
	aws := tofutestutils.AWS(t)

	log := logger.NewTestLogger(t)
	ctx := testutil.Context(t)

	testDir := t.TempDir()
	registryDir := path.Join(testDir, "registry")
	workDir := path.Join(testDir, "work")
	docsDir := path.Join(testDir, "docs")

	if err := os.MkdirAll(registryDir, 0755); err != nil {
		t.Fatalf("failed to create registry dir: %v", err)
	}
	if err := os.MkdirAll(workDir, 0755); err != nil {
		t.Fatalf("failed to create work dir: %v", err)
	}
	if err := os.MkdirAll(docsDir, 0755); err != nil {
		t.Fatalf("failed to create docs dir: %v", err)
	}

	s3Params := factory.S3Parameters{
		CACertFile: aws.CACertFile(),
		Bucket:     aws.S3Bucket(),
		Endpoint:   aws.S3Endpoint(),
		PathStyle:  aws.S3UsePathStyle(),
		AccessKey:  aws.AccessKey(),
		SecretKey:  aws.SecretKey(),
		Region:     aws.Region(),
	}

	backendFactory, err := factory.New(log, "")
	if err != nil {
		t.Fatalf("failed to create backend factory: %v", err)
	}
	binaryName := "tofu"
	if runtime.GOOS == "windows" {
		binaryName += ".exe"
	}
	// TODO this is only until these features get released in mainline tofu.
	tofuBinaryPath := path.Join("moduleindex", "moduleschema", "testtofu", binaryName)
	backendInstance, err := backendFactory.Create(ctx, registryDir, workDir, docsDir, blocklist.New(), s3Params, 25, tofuBinaryPath, testLicenseList)
	if err != nil {
		t.Fatalf("failed to create backend instance: %v", err)
	}

	if err := backendInstance.Generate(
		ctx,
		internal.WithNamespace("integrations"),
	); err != nil {
		t.Fatalf("failed to generate: %v", err)
	}

	// TODO check a few files if they are present on the S3 backend.
}

func TestE2EDoubleRun(t *testing.T) {
	aws := tofutestutils.AWS(t)

	log := logger.NewTestLogger(t)
	ctx := testutil.Context(t)

	testDir := t.TempDir()
	registryDir := path.Join(testDir, "registry")
	workDir := path.Join(testDir, "work")
	docsDir := path.Join(testDir, "docs")

	if err := os.MkdirAll(registryDir, 0755); err != nil {
		t.Fatalf("failed to create registry dir: %v", err)
	}
	if err := os.MkdirAll(workDir, 0755); err != nil {
		t.Fatalf("failed to create work dir: %v", err)
	}
	if err := os.MkdirAll(docsDir, 0755); err != nil {
		t.Fatalf("failed to create docs dir: %v", err)
	}

	s3Params := factory.S3Parameters{
		CACertFile: aws.CACertFile(),
		Bucket:     aws.S3Bucket(),
		Endpoint:   aws.S3Endpoint(),
		PathStyle:  aws.S3UsePathStyle(),
		AccessKey:  aws.AccessKey(),
		SecretKey:  aws.SecretKey(),
		Region:     aws.Region(),
	}

	backendFactory, err := factory.New(log, "")
	if err != nil {
		t.Fatalf("failed to create backend factory: %v", err)
	}
	binaryName := "tofu"
	if runtime.GOOS == "windows" {
		binaryName += ".exe"
	}
	// TODO this is only until these features get released in mainline tofu.
	tofuBinaryPath := path.Join("moduleindex", "moduleschema", "testtofu", binaryName)
	backendInstance, err := backendFactory.Create(ctx, registryDir, workDir, docsDir, blocklist.New(), s3Params, 25, tofuBinaryPath, testLicenseList)
	if err != nil {
		t.Fatalf("failed to create backend instance: %v", err)
	}

	t.Logf("🏃 Starting first run...")
	if err := backendInstance.Generate(
		ctx,
		internal.WithNamespace("integrations"),
	); err != nil {
		t.Fatalf("failed to generate (first run): %v", err)
	}

	// TODO check if the second run is actually using the files on the storage. Maybe via metrics collection?

	t.Logf("🏃 Starting second run...")
	if err := backendInstance.Generate(
		ctx,
		internal.WithNamespace("integrations"),
	); err != nil {
		t.Fatalf("failed to generate (second run): %v", err)
	}

	// TODO check a few files if they are present on the S3 backend.
}
