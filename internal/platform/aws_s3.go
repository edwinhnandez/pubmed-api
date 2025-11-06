package platform

import (
	"context"
	"fmt"
	"io"
	"log/slog"
	"strings"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/s3"
)

// LoadFromS3 loads data from an S3 bucket
func LoadFromS3(ctx context.Context, s3URL string, logger *slog.Logger) ([]byte, error) {
	// Parse S3 URL: s3://bucket/key
	if !strings.HasPrefix(s3URL, "s3://") {
		return nil, fmt.Errorf("invalid S3 URL format: %s", s3URL)
	}

	path := strings.TrimPrefix(s3URL, "s3://")
	parts := strings.SplitN(path, "/", 2)
	if len(parts) != 2 {
		return nil, fmt.Errorf("invalid S3 URL format: %s", s3URL)
	}

	bucket := parts[0]
	key := parts[1]

	logger.Info("loading data from S3", "bucket", bucket, "key", key)

	cfg, err := config.LoadDefaultConfig(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to load AWS config: %w", err)
	}

	client := s3.NewFromConfig(cfg)

	result, err := client.GetObject(ctx, &s3.GetObjectInput{
		Bucket: aws.String(bucket),
		Key:    aws.String(key),
	})
	if err != nil {
		return nil, fmt.Errorf("failed to get object from S3: %w", err)
	}
	defer result.Body.Close()

	data, err := io.ReadAll(result.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read S3 object: %w", err)
	}

	logger.Info("loaded data from S3", "size", len(data))
	return data, nil
}

