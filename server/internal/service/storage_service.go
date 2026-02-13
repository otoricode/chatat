package service

import (
	"context"
	"fmt"
	"io"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	awsconfig "github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/credentials"
	"github.com/aws/aws-sdk-go-v2/service/s3"

	"github.com/otoritech/chatat/internal/config"
)

// StorageService handles file operations with S3-compatible storage.
type StorageService interface {
	Upload(ctx context.Context, input UploadInput) (*UploadResult, error)
	GetURL(ctx context.Context, key string) (string, error)
	Delete(ctx context.Context, key string) error
}

// UploadInput holds parameters for uploading a file.
type UploadInput struct {
	Data        io.Reader
	Key         string // e.g., "media/chat/{chatId}/{uuid}.jpg"
	ContentType string
	Size        int64
}

// UploadResult holds the result of a file upload.
type UploadResult struct {
	Key  string `json:"key"`
	URL  string `json:"url"`
	Size int64  `json:"size"`
}

type s3StorageService struct {
	client       *s3.Client
	presigner    *s3.PresignClient
	bucket       string
	endpoint     string
}

// NewStorageService creates a new S3-compatible storage service.
func NewStorageService(cfg *config.Config) (StorageService, error) {
	resolver := aws.EndpointResolverWithOptionsFunc(
		func(svc, region string, options ...interface{}) (aws.Endpoint, error) {
			return aws.Endpoint{
				URL:               cfg.S3Endpoint,
				HostnameImmutable: true,
			}, nil
		},
	)

	awsCfg, err := awsconfig.LoadDefaultConfig(context.Background(),
		awsconfig.WithRegion(cfg.S3Region),
		awsconfig.WithCredentialsProvider(credentials.NewStaticCredentialsProvider(
			cfg.S3AccessKey, cfg.S3SecretKey, "",
		)),
		awsconfig.WithEndpointResolverWithOptions(resolver),
	)
	if err != nil {
		return nil, fmt.Errorf("loading AWS config: %w", err)
	}

	client := s3.NewFromConfig(awsCfg, func(o *s3.Options) {
		o.UsePathStyle = true // Required for MinIO
	})

	return &s3StorageService{
		client:    client,
		presigner: s3.NewPresignClient(client),
		bucket:    cfg.S3Bucket,
		endpoint:  cfg.S3Endpoint,
	}, nil
}

func (s *s3StorageService) Upload(ctx context.Context, input UploadInput) (*UploadResult, error) {
	_, err := s.client.PutObject(ctx, &s3.PutObjectInput{
		Bucket:      aws.String(s.bucket),
		Key:         aws.String(input.Key),
		Body:        input.Data,
		ContentType: aws.String(input.ContentType),
	})
	if err != nil {
		return nil, fmt.Errorf("uploading to S3: %w", err)
	}

	url := fmt.Sprintf("%s/%s/%s", s.endpoint, s.bucket, input.Key)

	return &UploadResult{
		Key:  input.Key,
		URL:  url,
		Size: input.Size,
	}, nil
}

func (s *s3StorageService) GetURL(ctx context.Context, key string) (string, error) {
	presigned, err := s.presigner.PresignGetObject(ctx, &s3.GetObjectInput{
		Bucket: aws.String(s.bucket),
		Key:    aws.String(key),
	}, func(po *s3.PresignOptions) {
		po.Expires = 1 * time.Hour
	})
	if err != nil {
		return "", fmt.Errorf("generating presigned URL: %w", err)
	}
	return presigned.URL, nil
}

func (s *s3StorageService) Delete(ctx context.Context, key string) error {
	_, err := s.client.DeleteObject(ctx, &s3.DeleteObjectInput{
		Bucket: aws.String(s.bucket),
		Key:    aws.String(key),
	})
	if err != nil {
		return fmt.Errorf("deleting from S3: %w", err)
	}
	return nil
}
