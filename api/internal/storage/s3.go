package storage

import (
	"context"
	"fmt"
	"io"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/credentials"
	"github.com/aws/aws-sdk-go-v2/service/s3"
	"github.com/klauspost/compress/zstd"
)

type S3Storage struct {
	client *s3.Client
	bucket string
}

func NewS3Storage(endpoint, accessKey, secretKey, bucket, region string) (*S3Storage, error) {
	cfg, err := config.LoadDefaultConfig(context.Background(),
		config.WithRegion(region),
		config.WithCredentialsProvider(credentials.NewStaticCredentialsProvider(accessKey, secretKey, "")),
	)
	if err != nil {
		return nil, fmt.Errorf("load aws config: %w", err)
	}

	client := s3.NewFromConfig(cfg, func(o *s3.Options) {
		if endpoint != "" {
			o.BaseEndpoint = aws.String(endpoint)
			o.UsePathStyle = true
		}
	})

	return &S3Storage{client: client, bucket: bucket}, nil
}

func (s *S3Storage) PresignUpload(ctx context.Context, deploymentID int) (string, error) {
	key := fmt.Sprintf("deployments/%d/output.zst", deploymentID)

	presignClient := s3.NewPresignClient(s.client)
	req, err := presignClient.PresignPutObject(ctx, &s3.PutObjectInput{
		Bucket: aws.String(s.bucket),
		Key:    aws.String(key),
	}, s3.WithPresignExpires(1*time.Hour))
	if err != nil {
		return "", fmt.Errorf("presign upload: %w", err)
	}

	return req.URL, nil
}

func (s *S3Storage) GetDeploymentOutput(ctx context.Context, deploymentID int) (string, error) {
	key := fmt.Sprintf("deployments/%d/output.zst", deploymentID)

	resp, err := s.client.GetObject(ctx, &s3.GetObjectInput{
		Bucket: aws.String(s.bucket),
		Key:    aws.String(key),
	})
	if err != nil {
		return "", fmt.Errorf("get object: %w", err)
	}
	defer func() { _ = resp.Body.Close() }()

	decoder, err := zstd.NewReader(resp.Body)
	if err != nil {
		return "", fmt.Errorf("create zstd decoder: %w", err)
	}
	defer decoder.Close()

	data, err := io.ReadAll(decoder)
	if err != nil {
		return "", fmt.Errorf("read decompressed data: %w", err)
	}

	return string(data), nil
}

func (s *S3Storage) DeleteDeploymentOutput(ctx context.Context, deploymentID int) error {
	key := fmt.Sprintf("deployments/%d/output.zst", deploymentID)

	_, err := s.client.DeleteObject(ctx, &s3.DeleteObjectInput{
		Bucket: aws.String(s.bucket),
		Key:    aws.String(key),
	})
	if err != nil {
		return fmt.Errorf("delete object: %w", err)
	}

	return nil
}
