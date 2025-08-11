package config

import (
	"context"
	"fmt"
	"log/slog"

	"github.com/aws/aws-sdk-go-v2/aws"
	awsconfig "github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/credentials"
	"github.com/aws/aws-sdk-go-v2/service/s3"
)

type BucketConfig struct {
	BucketName string `koanf:"name"`

	AccessKeyID     string `koanf:"accessKeyID"`
	SecretAccessKey string `koanf:"secretAccessKey"`
	Region          string `koanf:"region"`
	Endpoint        string `koanf:"endpoint"` // should always be auto for R2

	client *s3.Client
}

func (c *BucketConfig) validate() error {
	if c.AccessKeyID == "" {
		return fmt.Errorf("bucket.accessKeyID is required")
	}
	if c.SecretAccessKey == "" {
		return fmt.Errorf("bucket.secretAccessKey is required")
	}
	if c.Endpoint == "" {
		return fmt.Errorf("bucket.endpoint is required")
	}

	// Set default region if empty
	if c.Region == "" {
		c.Region = "auto"
	}

	return nil
}

func (c *BucketConfig) GetClient(ctx context.Context) (*s3.Client, error) {
	if c.client != nil {
		return c.client, nil
	}

	cfg, err := awsconfig.LoadDefaultConfig(ctx,
		awsconfig.WithCredentialsProvider(credentials.NewStaticCredentialsProvider(c.AccessKeyID, c.SecretAccessKey, "")),
		awsconfig.WithRegion("auto"),
	)
	if err != nil {
		slog.ErrorContext(ctx, "Failed to load AWS config", "error", err)
		return nil, err
	}

	client := s3.NewFromConfig(cfg, func(o *s3.Options) {
		o.BaseEndpoint = aws.String(c.Endpoint)
	})

	c.client = client

	return client, nil
}
