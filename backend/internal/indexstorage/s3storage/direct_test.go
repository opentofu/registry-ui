package s3storage_test

import (
	"context"
	"crypto/tls"
	"crypto/x509"
	"os"
	"testing"

	"github.com/opentofu/registry-ui/internal/indexstorage/s3storage"
	"github.com/opentofu/registry-ui/internal/testutil"
	"github.com/opentofu/tofutestutils"
)

func TestDirect(t *testing.T) {
	const testContent = "Hello world!"
	const testFile1 = "test.txt"
	const testDir = "test"
	const testFile2 = testDir + "/test.txt"

	aws := tofutestutils.AWS(t)
	certPool := x509.NewCertPool()
	certPool.AppendCertsFromPEM(aws.CACert())
	storage, err := s3storage.New(
		context.Background(),
		s3storage.WithBucket(aws.S3Bucket()),
		s3storage.WithAccessKey(aws.AccessKey()),
		s3storage.WithSecretKey(aws.SecretKey()),
		s3storage.WithRegion(aws.Region()),
		s3storage.WithEndpoint(aws.S3Endpoint()),
		s3storage.WithTLSConfig(&tls.Config{RootCAs: certPool}),
		s3storage.WithPathStyle(aws.S3UsePathStyle()),
	)
	if err != nil {
		t.Fatalf("failed to create S3 storage: %v", err)
	}

	t.Run("not-found", func(t *testing.T) {
		_, err := storage.ReadFile(testutil.Context(t), "non-existent.txt")
		if err == nil {
			t.Fatalf("Reading a non-existent file did not return an error")
		}
		if !os.IsNotExist(err) {
			t.Fatalf("Reading a non-existent file did not return an not-exists error (%T; %v)", err, err)
		}
	})

	t.Run("create", func(t *testing.T) {
		if err := storage.WriteFile(testutil.Context(t), testFile1, []byte(testContent)); err != nil {
			t.Fatalf("failed to write file: %v", err)
		}
	})

	t.Run("read", func(t *testing.T) {
		readContents, err := storage.ReadFile(testutil.Context(t), testFile1)
		if err != nil {
			t.Fatalf("failed to read file: %v", err)
		}
		if string(readContents) != testContent {
			t.Fatalf("Incorrect contents returned: %s", readContents)
		}
	})

	t.Run("create-subdir", func(t *testing.T) {
		if err := storage.WriteFile(testutil.Context(t), testFile2, []byte(testContent)); err != nil {
			t.Fatalf("failed to write file: %v", err)
		}
	})

	t.Run("read-subdir", func(t *testing.T) {
		readContents, err := storage.ReadFile(testutil.Context(t), testFile2)
		if err != nil {
			t.Fatalf("failed to read file: %v", err)
		}
		if string(readContents) != testContent {
			t.Fatalf("Incorrect contents returned: %s", readContents)
		}
	})

	t.Run("remove-subdir", func(t *testing.T) {
		if err := storage.RemoveAll(testutil.Context(t), testDir); err != nil {
			t.Fatalf("failed to remove all: %v", err)
		}
	})

	t.Run("read-subdir", func(t *testing.T) {
		_, err := storage.ReadFile(testutil.Context(t), testFile2)
		if err == nil {
			t.Fatalf("Reading a non-existent file did not return an error")
		}
		if !os.IsNotExist(err) {
			t.Fatalf("Reading a non-existent file did not return an not-exists error (%T; %v)", err, err)
		}
	})
}
