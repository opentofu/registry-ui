package s3storage_test

import (
	"crypto/tls"
	"crypto/x509"
	"os"
	"testing"

	"github.com/opentofu/registry-ui/internal/indexstorage/s3storage"
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
	storage := tofutestutils.Must2(s3storage.New(
		s3storage.WithBucket(aws.S3Bucket()),
		s3storage.WithAccessKey(aws.AccessKey()),
		s3storage.WithSecretKey(aws.SecretKey()),
		s3storage.WithRegion(aws.Region()),
		s3storage.WithEndpoint(aws.S3Endpoint()),
		s3storage.WithTLSConfig(&tls.Config{RootCAs: certPool}),
		s3storage.WithPathStyle(aws.S3UsePathStyle()),
	))

	t.Run("not-found", func(t *testing.T) {
		_, err := storage.ReadFile(tofutestutils.Context(t), "non-existent.txt")
		if err == nil {
			t.Fatalf("Reading a non-existent file did not return an error")
		}
		if !os.IsNotExist(err) {
			t.Fatalf("Reading a non-existent file did not return an not-exists error (%T; %v)", err, err)
		}
	})

	t.Run("create", func(t *testing.T) {
		tofutestutils.Must(storage.WriteFile(tofutestutils.Context(t), testFile1, []byte(testContent)))
	})

	t.Run("read", func(t *testing.T) {
		readContents := tofutestutils.Must2(storage.ReadFile(tofutestutils.Context(t), testFile1))
		if string(readContents) != testContent {
			t.Fatalf("Incorrect contents returned: %s", readContents)
		}
	})

	t.Run("create-subdir", func(t *testing.T) {
		tofutestutils.Must(storage.WriteFile(tofutestutils.Context(t), testFile2, []byte(testContent)))
	})

	t.Run("read-subdir", func(t *testing.T) {
		readContents := tofutestutils.Must2(storage.ReadFile(tofutestutils.Context(t), testFile2))
		if string(readContents) != testContent {
			t.Fatalf("Incorrect contents returned: %s", readContents)
		}
	})

	t.Run("remove-subdir", func(t *testing.T) {
		tofutestutils.Must(storage.RemoveAll(tofutestutils.Context(t), testDir))
	})

	t.Run("read-subdir", func(t *testing.T) {
		_, err := storage.ReadFile(tofutestutils.Context(t), testFile2)
		if err == nil {
			t.Fatalf("Reading a non-existent file did not return an error")
		}
		if !os.IsNotExist(err) {
			t.Fatalf("Reading a non-existent file did not return an not-exists error (%T; %v)", err, err)
		}
	})
}
