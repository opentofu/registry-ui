package index

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/feature/s3/manager"
	"github.com/aws/aws-sdk-go-v2/service/s3"
	"go.opentelemetry.io/otel/attribute"

	"github.com/opentofu/registry-ui/pkg/telemetry"
)

// UploadModuleVersionIndex uploads a module version index to S3
func UploadModuleVersionIndex(ctx context.Context, uploader *manager.Uploader, bucketName string, index *ModuleVersionIndex) error {
	key := fmt.Sprintf("modules/%s/%s/%s/index.json",
		index.Addr.Namespace, index.Addr.Name, index.Addr.Target)

	jsonData, err := json.MarshalIndent(index, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to marshal module index: %w", err)
	}

	return uploadToS3(ctx, uploader, bucketName, key, jsonData, "application/json")
}

// UploadProviderVersionIndex uploads a provider version index to S3
func UploadProviderVersionIndex(ctx context.Context, uploader *manager.Uploader, bucketName string, index *ProviderVersionIndex) error {
	key := fmt.Sprintf("providers/%s/%s/index.json",
		index.Addr.Namespace, index.Addr.Name)

	jsonData, err := json.MarshalIndent(index, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to marshal provider index: %w", err)
	}

	return uploadToS3(ctx, uploader, bucketName, key, jsonData, "application/json")
}

// uploadGlobalModuleIndex uploads the global module index to S3
func uploadGlobalModuleIndex(ctx context.Context, uploader *manager.Uploader, bucketName, key string, globalIndex *GlobalModuleIndex) error {
	jsonData, err := json.MarshalIndent(globalIndex, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to marshal global module index: %w", err)
	}

	return uploadToS3(ctx, uploader, bucketName, key, jsonData, "application/json")
}

// uploadGlobalProviderIndex uploads the global provider index to S3
func uploadGlobalProviderIndex(ctx context.Context, uploader *manager.Uploader, bucketName, key string, globalIndex *GlobalProviderIndex) error {
	jsonData, err := json.MarshalIndent(globalIndex, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to marshal global provider index: %w", err)
	}

	return uploadToS3(ctx, uploader, bucketName, key, jsonData, "application/json")
}

// uploadToS3 uploads data to S3 with the specified content type
func uploadToS3(ctx context.Context, uploader *manager.Uploader, bucketName, key string, data []byte, contentType string) error {
	// TODO: find a centralized way to upload files with decent OTEL in there, for now we'll just keep this method in this package
	ctx, span := telemetry.Tracer().Start(ctx, "index.upload_to_s3")
	defer span.End()

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
	)

	_, err := uploader.Upload(ctx, toUpload)
	if err != nil {
		span.RecordError(err)
		return err
	}
	return nil
}

// UploadGlobalModuleIndex is an exported wrapper for uploadGlobalModuleIndex
func UploadGlobalModuleIndex(ctx context.Context, uploader *manager.Uploader, bucketName, key string, globalIndex *GlobalModuleIndex) error {
	return uploadGlobalModuleIndex(ctx, uploader, bucketName, key, globalIndex)
}

// UploadGlobalProviderIndex is an exported wrapper for uploadGlobalProviderIndex
func UploadGlobalProviderIndex(ctx context.Context, uploader *manager.Uploader, bucketName, key string, globalIndex *GlobalProviderIndex) error {
	return uploadGlobalProviderIndex(ctx, uploader, bucketName, key, globalIndex)
}
