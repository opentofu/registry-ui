package storage

import (
	"bytes"
	"context"
	"crypto/md5"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"log/slog"
	"os"
	"path/filepath"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/feature/s3/manager"
	"github.com/aws/aws-sdk-go-v2/service/s3"
	"go.opentelemetry.io/otel/attribute"

	"github.com/opentofu/registry-ui/pkg/telemetry"
)

// StoreModuleInS3 stores module data in S3 and returns the MD5 checksum
func StoreModuleInS3(ctx context.Context, uploader *manager.Uploader, bucketName, namespace, name, target, version string, moduleData any) (string, error) {
	// Convert module data to JSON
	jsonData, err := json.MarshalIndent(moduleData, "", "  ")
	if err != nil {
		return "", fmt.Errorf("failed to marshal module data: %w", err)
	}

	key := fmt.Sprintf("modules/%s/%s/%s/%s/index.json", namespace, name, target, version)
	md5Hash, err := uploadToS3(ctx, uploader, bucketName, key, jsonData, "application/json")
	if err != nil {
		return "", fmt.Errorf("failed to upload module index.json: %w", err)
	}

	slog.DebugContext(ctx, "Stored module index.json in S3",
		"module", fmt.Sprintf("%s/%s/%s", namespace, name, target),
		"version", version,
		"key", key,
		"checksum", md5Hash)

	return md5Hash, nil
}

// StoreModuleSubmoduleInS3 stores submodule data and README in S3, returns (indexChecksum, readmeChecksum, error)
func StoreModuleSubmoduleInS3(ctx context.Context, uploader *manager.Uploader, bucketName, namespace, name, target, version, submoduleName string, tofuJSON any, workDir string) (string, string, error) {
	// Upload submodule index.json
	jsonData, err := json.MarshalIndent(tofuJSON, "", "  ")
	if err != nil {
		return "", "", fmt.Errorf("failed to marshal submodule data: %w", err)
	}

	indexKey := fmt.Sprintf("modules/%s/%s/%s/%s/submodules/%s/index.json", namespace, name, target, version, submoduleName)
	indexChecksum, err := uploadToS3(ctx, uploader, bucketName, indexKey, jsonData, "application/json")
	if err != nil {
		return "", "", fmt.Errorf("failed to upload submodule index.json: %w", err)
	}

	// Upload submodule README if it exists
	var readmeChecksum string
	readmePath := filepath.Join(workDir, "modules", submoduleName, "README.md")
	if readmeContent, err := os.ReadFile(readmePath); err == nil {
		readmeKey := fmt.Sprintf("modules/%s/%s/%s/%s/submodules/%s/README.md", namespace, name, target, version, submoduleName)
		readmeChecksum, err = uploadToS3(ctx, uploader, bucketName, readmeKey, readmeContent, "text/markdown")
		if err != nil {
			slog.WarnContext(ctx, "Failed to upload submodule README", "error", err, "submodule", submoduleName)
		} else {
			slog.DebugContext(ctx, "Uploaded submodule README", "key", readmeKey, "checksum", readmeChecksum)
		}
	}

	slog.DebugContext(ctx, "Stored submodule in S3",
		"module", fmt.Sprintf("%s/%s/%s", namespace, name, target),
		"version", version,
		"submodule", submoduleName,
		"index_checksum", indexChecksum,
		"readme_checksum", readmeChecksum)

	return indexChecksum, readmeChecksum, nil
}

// StoreModuleExampleInS3 stores example data and README in S3, returns (indexChecksum, readmeChecksum, error)
func StoreModuleExampleInS3(ctx context.Context, uploader *manager.Uploader, bucketName, namespace, name, target, version, exampleName string, tofuJSON any, workDir string) (string, string, error) {
	// Upload example index.json
	jsonData, err := json.MarshalIndent(tofuJSON, "", "  ")
	if err != nil {
		return "", "", fmt.Errorf("failed to marshal example data: %w", err)
	}

	indexKey := fmt.Sprintf("modules/%s/%s/%s/%s/examples/%s/index.json", namespace, name, target, version, exampleName)
	indexChecksum, err := uploadToS3(ctx, uploader, bucketName, indexKey, jsonData, "application/json")
	if err != nil {
		return "", "", fmt.Errorf("failed to upload example index.json: %w", err)
	}

	// Upload example README if it exists
	var readmeChecksum string
	readmePath := filepath.Join(workDir, "examples", exampleName, "README.md")
	if readmeContent, err := os.ReadFile(readmePath); err == nil {
		readmeKey := fmt.Sprintf("modules/%s/%s/%s/%s/examples/%s/README.md", namespace, name, target, version, exampleName)
		readmeChecksum, err = uploadToS3(ctx, uploader, bucketName, readmeKey, readmeContent, "text/markdown")
		if err != nil {
			slog.WarnContext(ctx, "Failed to upload example README", "error", err, "example", exampleName)
		} else {
			slog.DebugContext(ctx, "Uploaded example README", "key", readmeKey, "checksum", readmeChecksum)
		}
	}

	slog.DebugContext(ctx, "Stored example in S3",
		"module", fmt.Sprintf("%s/%s/%s", namespace, name, target),
		"version", version,
		"example", exampleName,
		"index_checksum", indexChecksum,
		"readme_checksum", readmeChecksum)

	return indexChecksum, readmeChecksum, nil
}

// StoreModuleREADME stores the main module README in S3 and returns the MD5 checksum
func StoreModuleREADME(ctx context.Context, uploader *manager.Uploader, bucketName, namespace, name, target, version, workDir string) (string, error) {
	readmePath := filepath.Join(workDir, "README.md")
	readmeContent, err := os.ReadFile(readmePath)
	if err != nil {
		if os.IsNotExist(err) {
			slog.DebugContext(ctx, "No README.md found for module",
				"module", fmt.Sprintf("%s/%s/%s", namespace, name, target))
			return "", nil // Not an error - just no README
		}
		return "", fmt.Errorf("failed to read README.md: %w", err)
	}

	readmeKey := fmt.Sprintf("modules/%s/%s/%s/%s/README.md", namespace, name, target, version)
	md5Hash, err := uploadToS3(ctx, uploader, bucketName, readmeKey, readmeContent, "text/markdown")
	if err != nil {
		return "", fmt.Errorf("failed to upload module README: %w", err)
	}

	slog.DebugContext(ctx, "Stored module README in S3",
		"module", fmt.Sprintf("%s/%s/%s", namespace, name, target),
		"version", version,
		"key", readmeKey,
		"checksum", md5Hash)

	return md5Hash, nil
}

// uploadToS3 uploads data to S3 with the specified content type and returns the MD5 checksum
func uploadToS3(ctx context.Context, uploader *manager.Uploader, bucketName, key string, data []byte, contentType string) (string, error) {
	// TODO: find a centralized way to upload files with decent OTEL in there, for now we'll just keep this method in this package
	ctx, span := telemetry.Tracer().Start(ctx, "module_storage.upload_to_s3")
	defer span.End()

	hash := md5.Sum(data)
	checksum := hex.EncodeToString(hash[:])

	toUpload := &s3.PutObjectInput{
		Bucket:      aws.String(bucketName),
		Key:         aws.String(key),
		Body:        bytes.NewReader(data),
		ContentType: aws.String(contentType),
	}

	span.SetAttributes(attribute.String("bucket", bucketName),
		attribute.String("key", key),
		attribute.String("content-type", contentType),
		attribute.Int64("size", int64(len(data))),
		attribute.String("md5_checksum", checksum),
	)

	_, err := uploader.Upload(ctx, toUpload)
	if err != nil {
		span.RecordError(err)
		return "", err
	}

	slog.DebugContext(ctx, "Uploaded to S3", "key", key, "checksum", checksum)
	return checksum, nil
}
