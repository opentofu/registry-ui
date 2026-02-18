package config

import (
	"context"
	"fmt"
	"log/slog"
	"net"
	"net/http"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	awsconfig "github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/credentials"
	"github.com/aws/aws-sdk-go-v2/service/s3"

	"go.opentelemetry.io/contrib/instrumentation/github.com/aws/aws-sdk-go-v2/otelaws"
	"go.opentelemetry.io/contrib/instrumentation/net/http/otelhttp"
)

type BucketConfig struct {
	BucketName string `koanf:"name"`

	AccessKeyID     string `koanf:"accesskeyid"`
	SecretAccessKey string `koanf:"secretaccesskey"`
	Region          string `koanf:"region"`
	Endpoint        string `koanf:"endpoint"` // S3-compatible endpoint URL, leave empty for AWS S3

	client     *s3.Client
	httpClient *http.Client
}

func (c *BucketConfig) Validate() error {
	if c.AccessKeyID == "" {
		return fmt.Errorf("bucket.accessKeyID is required")
	}
	if c.SecretAccessKey == "" {
		return fmt.Errorf("bucket.secretAccessKey is required")
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

	// Configure transport for high-concurrency uploads with connection reuse
	transport := &http.Transport{
		// Connection pool settings - critical for performance
		MaxIdleConns:        200,              // Total max idle connections
		MaxIdleConnsPerHost: 200,              // Max idle connections per host (default is only 2!)
		MaxConnsPerHost:     0,                // No limit on connections per host
		IdleConnTimeout:     90 * time.Second, // Keep connections alive longer

		// Timeouts
		DialContext: (&net.Dialer{
			Timeout:   30 * time.Second,
			KeepAlive: 30 * time.Second, // Enable TCP keep-alive
		}).DialContext,
		TLSHandshakeTimeout:   10 * time.Second,
		ResponseHeaderTimeout: 30 * time.Second,
		ExpectContinueTimeout: 1 * time.Second,

		// Enable HTTP/2 for better multiplexing (if supported by R2)
		ForceAttemptHTTP2: true,
	}

	httpClient := &http.Client{
		Transport: otelhttp.NewTransport(transport),
		Timeout:   60 * time.Second,
	}
	c.httpClient = httpClient

	cfg, err := awsconfig.LoadDefaultConfig(ctx,
		awsconfig.WithCredentialsProvider(credentials.NewStaticCredentialsProvider(c.AccessKeyID, c.SecretAccessKey, "")),
		awsconfig.WithRegion(c.Region),
		awsconfig.WithHTTPClient(httpClient),
	)
	if err != nil {
		slog.ErrorContext(ctx, "Failed to load AWS config", "error", err)
		return nil, err
	}

	// Add OTEL tracing for all s3 calls
	otelaws.AppendMiddlewares(&cfg.APIOptions)

	client := s3.NewFromConfig(cfg, func(o *s3.Options) {
		if c.Endpoint != "" {
			o.BaseEndpoint = aws.String(c.Endpoint)
		}
	})

	c.client = client

	slog.InfoContext(ctx, "S3 client initialized with optimized connection pooling",
		"max_idle_conns", 200,
		"max_idle_conns_per_host", 200,
		"endpoint", c.Endpoint)

	return client, nil
}
