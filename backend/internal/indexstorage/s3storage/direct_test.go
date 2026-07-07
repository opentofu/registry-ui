package s3storage_test

import (
	"context"
	"crypto/tls"
	"crypto/x509"
	"net/http/httptest"
	"os"
	"testing"

	"github.com/johannesboyne/gofakes3"
	"github.com/johannesboyne/gofakes3/backend/s3mem"
	"github.com/opentofu/registry-ui/internal/indexstorage"
	"github.com/opentofu/registry-ui/internal/indexstorage/s3storage"
)

func TestDirect(t *testing.T) {
	const testContent = "Hello world!"
	const testFile1 = "test.txt"
	const testDir = "test"
	const testFile2 = testDir + "/test.txt"

	storage := newTestStorage(t)

	t.Run("not-found", func(t *testing.T) {
		_, err := storage.ReadFile(t.Context(), "non-existent.txt")
		if err == nil {
			t.Fatalf("Reading a non-existent file did not return an error")
		}
		if !os.IsNotExist(err) {
			t.Fatalf("Reading a non-existent file did not return an not-exists error (%T; %v)", err, err)
		}
	})

	t.Run("create", func(t *testing.T) {
		if err := storage.WriteFile(t.Context(), testFile1, []byte(testContent)); err != nil {
			t.Fatalf("failed to write file: %v", err)
		}
	})

	t.Run("read", func(t *testing.T) {
		readContents, err := storage.ReadFile(t.Context(), testFile1)
		if err != nil {
			t.Fatalf("failed to read file: %v", err)
		}
		if string(readContents) != testContent {
			t.Fatalf("Incorrect contents returned: %s", readContents)
		}
	})

	t.Run("create-subdir", func(t *testing.T) {
		if err := storage.WriteFile(t.Context(), testFile2, []byte(testContent)); err != nil {
			t.Fatalf("failed to write file: %v", err)
		}
	})

	t.Run("read-subdir", func(t *testing.T) {
		readContents, err := storage.ReadFile(t.Context(), testFile2)
		if err != nil {
			t.Fatalf("failed to read file: %v", err)
		}
		if string(readContents) != testContent {
			t.Fatalf("Incorrect contents returned: %s", readContents)
		}
	})

	t.Run("remove-subdir", func(t *testing.T) {
		if err := storage.RemoveAll(t.Context(), testDir); err != nil {
			t.Fatalf("failed to remove all: %v", err)
		}
	})

	t.Run("read-subdir", func(t *testing.T) {
		_, err := storage.ReadFile(t.Context(), testFile2)
		if err == nil {
			t.Fatalf("Reading a non-existent file did not return an error")
		}
		if !os.IsNotExist(err) {
			t.Fatalf("Reading a non-existent file did not return an not-exists error (%T; %v)", err, err)
		}
	})
}

func newTestStorage(t *testing.T) indexstorage.API {
	t.Helper()

	const bucket = "test-bucket"
	backing := s3mem.New()
	if err := backing.CreateBucket(bucket); err != nil {
		t.Fatalf("failed to create bucket: %v", err)
	}

	server := httptest.NewTLSServer(gofakes3.New(backing).Server())
	t.Cleanup(server.Close)

	certPool := x509.NewCertPool()
	certPool.AddCert(server.Certificate())

	storage, err := s3storage.New(
		context.Background(),
		s3storage.WithBucket(bucket),
		s3storage.WithAccessKey("test"),
		s3storage.WithSecretKey("test"),
		s3storage.WithRegion("us-east-1"),
		s3storage.WithEndpoint(server.URL),
		s3storage.WithTLSConfig(&tls.Config{RootCAs: certPool}),
		s3storage.WithPathStyle(true),
	)
	if err != nil {
		t.Fatalf("failed to create S3 storage: %v", err)
	}
	return storage
}
