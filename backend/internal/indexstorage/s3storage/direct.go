package s3storage

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"io"
	"io/fs"
	"strings"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/s3"
	"github.com/aws/aws-sdk-go-v2/service/s3/types"
	"github.com/opentofu/registry-ui/internal/indexstorage"
)

// New creates an API instance that uses an S3 backend as a storage.
func New(ctx context.Context, opts ...Opt) (indexstorage.API, error) {
	cfg := Config{}
	for _, opt := range opts {
		if err := opt(&cfg); err != nil {
			return nil, err
		}
	}
	if err := cfg.applyDefaults(ctx); err != nil {
		return nil, err
	}
	if err := cfg.Validate(ctx); err != nil {
		return nil, err
	}
	awsConfig := cfg.awsConfig()
	s3Connection := s3.NewFromConfig(awsConfig, func(options *s3.Options) {
		options.UsePathStyle = cfg.PathStyle
	})
	return &directAPI{
		cfg,
		s3Connection,
		"",
	}, nil
}

type directAPI struct {
	cfg    Config
	client *s3.Client
	prefix indexstorage.Path
}

func (d directAPI) ReadFile(ctx context.Context, objectPath indexstorage.Path) ([]byte, error) {
	finalPath := d.prefix + objectPath
	if err := finalPath.Validate(); err != nil {
		return nil, fmt.Errorf("invalid path %s (%w)", finalPath, err)
	}
	d.cfg.logger.Trace(ctx, "Downloading %s ...", finalPath)
	object, err := d.client.GetObject(ctx, &s3.GetObjectInput{
		Bucket: aws.String(d.cfg.Bucket),
		Key:    aws.String(string(finalPath)),
	})
	if err != nil {
		var notFoundError *types.NoSuchKey
		if errors.As(err, &notFoundError) {
			// TODO the higher level implementations rely on os.IsNotExist(), which is not satisfied by wrapped errors.
			d.cfg.logger.Trace(ctx, "Object %s does not exist (%v). This is normal, don't worry.", finalPath, err)
			return nil, fs.ErrNotExist
		}
		d.cfg.logger.Warn(ctx, "Downloading %s failed (%v).", finalPath, err)
		return nil, fmt.Errorf("failed to request %s (%w)", finalPath, err)
	}
	body := object.Body
	defer func() {
		_ = body.Close()
	}()
	data, err := io.ReadAll(body)
	if err != nil {
		d.cfg.logger.Warn(ctx, "Downloading %s failed (%v).", finalPath, err)
		return nil, fmt.Errorf("failed to read %s (%w)", finalPath, err)
	}
	d.cfg.logger.Trace(ctx, "Downloaded %s.", finalPath)
	return data, nil
}

func (d directAPI) WriteFile(ctx context.Context, objectPath indexstorage.Path, contents []byte) error {
	finalPath := d.prefix + objectPath
	if err := finalPath.Validate(); err != nil {
		return fmt.Errorf("invalid path %s (%w)", finalPath, err)
	}
	d.cfg.logger.Trace(ctx, "Uploading %s ...", finalPath)
	contentType := "application/octet-stream"
	if strings.HasSuffix(string(objectPath), ".html") {
		contentType = "text/html"
	} else if strings.HasSuffix(string(objectPath), ".json") {
		contentType = "application/json"
	} else if strings.HasSuffix(string(objectPath), ".md") {
		contentType = "text/markdown"
	}

	if _, err := d.client.PutObject(ctx, &s3.PutObjectInput{
		Bucket:      aws.String(d.cfg.Bucket),
		Key:         aws.String(string(finalPath)),
		Body:        bytes.NewReader(contents),
		ContentType: aws.String(contentType),
	}); err != nil {
		d.cfg.logger.Warn(ctx, "Uploading %s failed (%v).", finalPath, err)
		return fmt.Errorf("failed to put %s (%w)", objectPath, err)
	}
	d.cfg.logger.Trace(ctx, "Uploaded %s.", finalPath)
	return nil
}

func (d directAPI) RemoveAll(ctx context.Context, objectPath indexstorage.Path) error {
	finalPath := d.prefix + objectPath
	if err := finalPath.Validate(); err != nil {
		return fmt.Errorf("invalid path %s (%w)", finalPath, err)
	}

	d.cfg.logger.Trace(ctx, "Deleting %s/* ...", string(finalPath))
	var continuationToken *string
	i := 1
	for {
		d.cfg.logger.Trace(ctx, "Fetching object list for %s [%d] ...", string(finalPath), i)
		listResponse, err := d.client.ListObjectsV2(ctx, &s3.ListObjectsV2Input{
			Bucket:            aws.String(d.cfg.Bucket),
			ContinuationToken: continuationToken,
			Prefix:            aws.String(string(finalPath) + "/"),
		})
		if err != nil {
			d.cfg.logger.Warn(ctx, "Fetching object list for %s [%d] failed (%v).", string(finalPath), i, err)
			return fmt.Errorf("failed to list objects with prefix %s (%w)", objectPath, err)
		}
		if len(listResponse.Contents) == 0 {
			d.cfg.logger.Trace(ctx, "No objects to delete under %s.", string(finalPath))
			return nil
		}
		objects := make([]types.ObjectIdentifier, len(listResponse.Contents))
		for i, object := range listResponse.Contents {
			objects[i] = types.ObjectIdentifier{
				Key: object.Key,
			}
		}
		d.cfg.logger.Trace(ctx, "Deleting %d object(s) matching %s/* [%d] ...", finalPath, i)
		if _, err := d.client.DeleteObjects(ctx, &s3.DeleteObjectsInput{
			Bucket: aws.String(d.cfg.Bucket),
			Delete: &types.Delete{
				Objects: objects,
			},
			BypassGovernanceRetention: nil,
			ChecksumAlgorithm:         "",
			ExpectedBucketOwner:       nil,
			MFA:                       nil,
			RequestPayer:              "",
		}); err != nil {
			d.cfg.logger.Trace(ctx, "Deleting %d object(s) matching %s/* [%d] failed (%v)", finalPath, i, err)
			return fmt.Errorf("failed to delete objects (%w)", err)
		}

		if listResponse.NextContinuationToken == nil {
			break
		}
		continuationToken = listResponse.NextContinuationToken
		i++
	}
	d.cfg.logger.Trace(ctx, "Completed deleting %s/*.", finalPath)
	return nil
}

func (d directAPI) Subdirectory(_ context.Context, dir indexstorage.Path) (indexstorage.API, error) {
	newPrefix := d.prefix + dir + "/"
	if err := newPrefix.Validate(); err != nil {
		return nil, err
	}
	return &directAPI{
		cfg:    d.cfg,
		client: d.client,
		prefix: newPrefix,
	}, nil
}
