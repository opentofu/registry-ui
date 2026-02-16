package index

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/feature/s3/manager"
	"github.com/aws/aws-sdk-go-v2/service/s3"
	"github.com/aws/aws-sdk-go-v2/service/s3/types"
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

// UpdateGlobalModuleIndex downloads the global module index, updates it with the new module, and uploads it back
func UpdateGlobalModuleIndex(ctx context.Context, s3Client *s3.Client, uploader *manager.Uploader, bucketName string, moduleIndex *ModuleVersionIndex) error {
	key := "modules/index.json"

	// 1. Download existing global index from S3
	globalIndex, err := downloadGlobalModuleIndex(ctx, s3Client, bucketName, key)
	if err != nil {
		return fmt.Errorf("failed to download global module index: %w", err)
	}

	// 2. Update/add entry for this module
	if len(moduleIndex.Versions) > 0 {
		// Get published date, defaulting to zero time if nil
		publishedAt := time.Time{}
		if moduleIndex.Versions[0].Published != nil {
			publishedAt = *moduleIndex.Versions[0].Published
		}

		moduleEntry := ModuleEntry{
			Addr:          moduleIndex.Addr,
			Description:   moduleIndex.Description,
			LatestVersion: moduleIndex.Versions[0].ID, // assuming sorted by newest first
			PublishedAt:   publishedAt,
		}
		globalIndex.UpdateOrAdd(moduleEntry)
	}

	// 3. Upload back to S3
	return uploadGlobalModuleIndex(ctx, uploader, bucketName, key, globalIndex)
}

// UpdateGlobalProviderIndex downloads the global provider index, updates it with the new provider, and uploads it back
func UpdateGlobalProviderIndex(ctx context.Context, s3Client *s3.Client, uploader *manager.Uploader, bucketName string, providerIndex *ProviderVersionIndex) error {
	key := "providers/index.json"

	// 1. Download existing global index from S3
	globalIndex, err := downloadGlobalProviderIndex(ctx, s3Client, bucketName, key)
	if err != nil {
		return fmt.Errorf("failed to download global provider index: %w", err)
	}

	// 2. Update/add entry for this provider
	if len(providerIndex.Versions) > 0 {
		// Generate GitHub repository URL
		repoURL := fmt.Sprintf("https://github.com/%s/terraform-provider-%s",
			providerIndex.Addr.Namespace, providerIndex.Addr.Name)

		// Get published date, defaulting to zero time if nil
		publishedAt := time.Time{}
		if providerIndex.Versions[0].Published != nil {
			publishedAt = *providerIndex.Versions[0].Published
		}

		providerEntry := ProviderEntry{
			Addr:          providerIndex.Addr,
			Description:   providerIndex.Description,
			LatestVersion: providerIndex.Versions[0].ID, // assuming sorted by newest first
			PublishedAt:   publishedAt,
			Link:          repoURL,
			Warnings:      providerIndex.Warnings,
			Popularity:    providerIndex.Popularity, // Using GitHub stars as popularity
			ForkCount:     providerIndex.ForkCount,
			ForkOf:        providerIndex.ForkOf,
			ForkOfLink:    providerIndex.ForkOfLink,
		}
		globalIndex.UpdateOrAdd(providerEntry)
	}

	// 3. Upload back to S3
	return uploadGlobalProviderIndex(ctx, uploader, bucketName, key, globalIndex)
}

// downloadGlobalModuleIndex downloads and parses the global module index, or returns empty index if not found
func downloadGlobalModuleIndex(ctx context.Context, s3Client *s3.Client, bucketName, key string) (*GlobalModuleIndex, error) {
	result, err := s3Client.GetObject(ctx, &s3.GetObjectInput{
		Bucket: aws.String(bucketName),
		Key:    aws.String(key),
	})
	if err != nil {
		// If object doesn't exist, return empty index
		var noSuchKey *types.NoSuchKey
		if ok := errors.As(err, &noSuchKey); ok {
			return &GlobalModuleIndex{Modules: []ModuleEntry{}}, nil
		}
		return nil, fmt.Errorf("failed to get global module index from S3: %w", err)
	}
	defer result.Body.Close()

	data, err := io.ReadAll(result.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read global module index: %w", err)
	}

	var globalIndex GlobalModuleIndex
	err = json.Unmarshal(data, &globalIndex)
	if err != nil {
		// If JSON is invalid, return empty index
		return &GlobalModuleIndex{Modules: []ModuleEntry{}}, nil
	}

	return &globalIndex, nil
}

// downloadGlobalProviderIndex downloads and parses the global provider index, or returns empty index if not found
func downloadGlobalProviderIndex(ctx context.Context, s3Client *s3.Client, bucketName, key string) (*GlobalProviderIndex, error) {
	result, err := s3Client.GetObject(ctx, &s3.GetObjectInput{
		Bucket: aws.String(bucketName),
		Key:    aws.String(key),
	})
	if err != nil {
		// If object doesn't exist, return empty index
		var noSuchKey *types.NoSuchKey
		if ok := errors.As(err, &noSuchKey); ok {
			return &GlobalProviderIndex{Providers: []ProviderEntry{}}, nil
		}
		return nil, fmt.Errorf("failed to get global provider index from S3: %w", err)
	}
	defer result.Body.Close()

	data, err := io.ReadAll(result.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read global provider index: %w", err)
	}

	var globalIndex GlobalProviderIndex
	err = json.Unmarshal(data, &globalIndex)
	if err != nil {
		// If JSON is invalid, return empty index
		return &GlobalProviderIndex{Providers: []ProviderEntry{}}, nil
	}

	return &globalIndex, nil
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
	ctx, span := telemetry.Tracer().Start(ctx, "upload-to-s3")
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

// UpdateOrAdd updates an existing module entry or adds a new one to the global index
func (g *GlobalModuleIndex) UpdateOrAdd(entry ModuleEntry) {
	// Find existing entry
	for i, existing := range g.Modules {
		if existing.Addr.Namespace == entry.Addr.Namespace &&
			existing.Addr.Name == entry.Addr.Name &&
			existing.Addr.Target == entry.Addr.Target {
			// Update existing entry
			g.Modules[i] = entry
			return
		}
	}
	// Add new entry
	g.Modules = append(g.Modules, entry)
}

// UpdateOrAdd updates an existing provider entry or adds a new one to the global index
func (g *GlobalProviderIndex) UpdateOrAdd(entry ProviderEntry) {
	// Find existing entry
	for i, existing := range g.Providers {
		if existing.Addr.Namespace == entry.Addr.Namespace &&
			existing.Addr.Name == entry.Addr.Name {
			// Update existing entry
			g.Providers[i] = entry
			return
		}
	}
	// Add new entry
	g.Providers = append(g.Providers, entry)
}

// UploadGlobalModuleIndex is an exported wrapper for uploadGlobalModuleIndex
func UploadGlobalModuleIndex(ctx context.Context, uploader *manager.Uploader, bucketName, key string, globalIndex *GlobalModuleIndex) error {
	return uploadGlobalModuleIndex(ctx, uploader, bucketName, key, globalIndex)
}

// UploadGlobalProviderIndex is an exported wrapper for uploadGlobalProviderIndex
func UploadGlobalProviderIndex(ctx context.Context, uploader *manager.Uploader, bucketName, key string, globalIndex *GlobalProviderIndex) error {
	return uploadGlobalProviderIndex(ctx, uploader, bucketName, key, globalIndex)
}
