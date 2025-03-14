package s3storage

import (
	"context"
	"crypto/tls"
	"fmt"
	"log/slog"
	"net/http"

	awsv2 "github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/s3"
)

type Config struct {
	Bucket    string
	Region    string
	Endpoint  string
	PathStyle bool
	AccessKey string
	SecretKey string
	TLSConfig *tls.Config
	logger    *slog.Logger
}

func (c *Config) Validate(ctx context.Context) error {
	if c.AccessKey == "" {
		return fmt.Errorf("the AWS access key is required")
	}
	if c.SecretKey == "" {
		return fmt.Errorf("the AWS secret key is required")
	}
	if c.Bucket == "" {
		return fmt.Errorf("the bucket is required")
	}

	c.logger.DebugContext(ctx, "Checking S3 connection...")
	_, err := c.awsClient().ListObjectsV2(ctx, &s3.ListObjectsV2Input{
		Bucket:  awsv2.String(c.Bucket),
		MaxKeys: awsv2.Int32(1),
	})
	if err != nil {
		c.logger.ErrorContext(ctx, "S3 connection failed (%v)", err)
		return fmt.Errorf("incorrect S3 parameters (%w)", err)
	}
	c.logger.DebugContext(ctx, "S3 connection successful.")

	return nil
}

func (c *Config) applyDefaults(ctx context.Context) error {
	c.logger = c.logger.With(slog.String("name", "S3"))
	if c.Region == "" {
		c.logger.DebugContext(ctx, "No bucket location specified, attempting to automatically determine location...")
		cli := c.awsClient()
		locationResponse, err := cli.GetBucketLocation(ctx, &s3.GetBucketLocationInput{
			Bucket: awsv2.String(c.Bucket),
		})
		if err != nil {
			c.logger.ErrorContext(ctx, "GetBucketLocation query failed (%v)", err)
			return fmt.Errorf("no bucket region provided and automatic region lookup failed (%w)", err)
		}
		if locationResponse.LocationConstraint == "" {
			c.Region = "us-east-1"
		} else {
			c.Region = string(locationResponse.LocationConstraint)
		}
	}
	return nil
}

func (c *Config) awsClient() *s3.Client {
	awsConfig := c.awsConfig()
	return s3.NewFromConfig(awsConfig, func(options *s3.Options) {
		options.UsePathStyle = c.PathStyle
	})
}

func (c *Config) awsConfig() awsv2.Config {
	var endpoint *string
	if c.Endpoint != "" {
		endpoint = awsv2.String(c.Endpoint)
	}
	region := c.Region
	if region == "" {
		region = "us-east-1"
	}

	return awsv2.Config{
		Region: region,
		Credentials: awsv2.CredentialsProviderFunc(func(_ context.Context) (awsv2.Credentials, error) {
			return awsv2.Credentials{
				AccessKeyID:     c.AccessKey,
				SecretAccessKey: c.SecretKey,
			}, nil
		}),
		BaseEndpoint: endpoint,
		HTTPClient: &http.Client{
			Transport: &http.Transport{
				TLSClientConfig: c.TLSConfig,
			},
		},
	}
}

type Opt func(config *Config) error

func WithLogger(log *slog.Logger) Opt {
	return func(config *Config) error {
		config.logger = log
		return nil
	}
}

func WithBucket(bucket string) Opt {
	return func(config *Config) error {
		config.Bucket = bucket
		return nil
	}
}

func WithEndpoint(endpoint string) Opt {
	return func(config *Config) error {
		config.Endpoint = endpoint
		return nil
	}
}

func WithRegion(region string) Opt {
	return func(config *Config) error {
		config.Region = region
		return nil
	}
}

func WithPathStyle(pathStyle bool) Opt {
	return func(config *Config) error {
		config.PathStyle = pathStyle
		return nil
	}
}

func WithAccessKey(accessKey string) Opt {
	return func(config *Config) error {
		config.AccessKey = accessKey
		return nil
	}
}

func WithSecretKey(secretKey string) Opt {
	return func(config *Config) error {
		config.SecretKey = secretKey
		return nil
	}
}

func WithTLSConfig(tlsConfig *tls.Config) Opt {
	return func(config *Config) error {
		config.TLSConfig = tlsConfig
		return nil
	}
}
