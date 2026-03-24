package service

import (
	"context"
	"errors"
	"fmt"
	"io"
	"net/url"
	"strings"

	"github.com/aws/aws-sdk-go-v2/aws"
	awsconfig "github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/credentials"
	"github.com/aws/aws-sdk-go-v2/service/s3"
	"github.com/google/uuid"

	"maid-recruitment-tracking/internal/config"
)

var ErrUnsupportedContentType = errors.New("unsupported content type")

type StorageService interface {
	Upload(file io.Reader, fileName, contentType string) (url string, err error)
	Delete(url string) error
}

type S3StorageService struct {
	client        *s3.Client
	bucket        string
	region        string
	endpoint      string
	publicBaseURL string
}

var allowedContentTypes = map[string]struct{}{
	"image/jpeg":      {},
	"image/png":       {},
	"video/mp4":       {},
	"application/pdf": {},
}

func NewS3StorageService(cfg *config.Config) (*S3StorageService, error) {
	if cfg == nil {
		return nil, fmt.Errorf("config is nil")
	}
	if strings.TrimSpace(cfg.S3Bucket) == "" {
		return nil, fmt.Errorf("s3 bucket is empty")
	}
	if strings.TrimSpace(cfg.AWSRegion) == "" {
		return nil, fmt.Errorf("aws region is empty")
	}
	if strings.TrimSpace(cfg.AWSAccessKey) == "" || strings.TrimSpace(cfg.AWSSecretKey) == "" {
		return nil, fmt.Errorf("aws credentials are missing")
	}

	awsCfg, err := awsconfig.LoadDefaultConfig(context.Background(),
		awsconfig.WithRegion(cfg.AWSRegion),
		awsconfig.WithCredentialsProvider(credentials.NewStaticCredentialsProvider(cfg.AWSAccessKey, cfg.AWSSecretKey, "")),
	)
	if err != nil {
		return nil, fmt.Errorf("load aws config: %w", err)
	}

	client := s3.NewFromConfig(awsCfg, func(o *s3.Options) {
		if strings.TrimSpace(cfg.S3Endpoint) != "" {
			o.BaseEndpoint = aws.String(cfg.S3Endpoint)
			o.UsePathStyle = true
		}
	})

	return &S3StorageService{
		client:        client,
		bucket:        cfg.S3Bucket,
		region:        cfg.AWSRegion,
		endpoint:      cfg.S3Endpoint,
		publicBaseURL: strings.TrimRight(strings.TrimSpace(cfg.S3PublicBaseURL), "/"),
	}, nil
}

func (s *S3StorageService) Upload(file io.Reader, fileName, contentType string) (string, error) {
	if _, ok := allowedContentTypes[contentType]; !ok {
		return "", ErrUnsupportedContentType
	}
	if file == nil {
		return "", fmt.Errorf("file is nil")
	}

	ext := extensionForContentType(contentType)
	if ext == "" {
		return "", ErrUnsupportedContentType
	}
	objectKey := fmt.Sprintf("documents/%s%s", uuid.NewString(), ext)

	_, err := s.client.PutObject(context.Background(), &s3.PutObjectInput{
		Bucket:      aws.String(s.bucket),
		Key:         aws.String(objectKey),
		Body:        file,
		ContentType: aws.String(contentType),
	})
	if err != nil {
		return "", fmt.Errorf("upload to s3: %w", err)
	}

	return s.publicURL(objectKey), nil
}

func (s *S3StorageService) Delete(fileURL string) error {
	key, err := s.extractObjectKey(fileURL)
	if err != nil {
		return err
	}

	_, err = s.client.DeleteObject(context.Background(), &s3.DeleteObjectInput{
		Bucket: aws.String(s.bucket),
		Key:    aws.String(key),
	})
	if err != nil {
		return fmt.Errorf("delete from s3: %w", err)
	}
	return nil
}

func (s *S3StorageService) publicURL(key string) string {
	if strings.TrimSpace(s.publicBaseURL) != "" {
		return s.publicBaseURL + "/" + key
	}
	if strings.TrimSpace(s.endpoint) != "" {
		return strings.TrimRight(s.endpoint, "/") + "/" + s.bucket + "/" + key
	}
	return fmt.Sprintf("https://%s.s3.%s.amazonaws.com/%s", s.bucket, s.region, key)
}

func (s *S3StorageService) extractObjectKey(fileURL string) (string, error) {
	parsedURL, err := url.Parse(fileURL)
	if err != nil {
		return "", fmt.Errorf("parse url: %w", err)
	}

	trimmedPath := strings.TrimPrefix(parsedURL.Path, "/")
	if strings.TrimSpace(s.publicBaseURL) != "" {
		publicBase, err := url.Parse(s.publicBaseURL)
		if err == nil {
			basePath := strings.Trim(strings.TrimSpace(publicBase.Path), "/")
			if basePath != "" && strings.HasPrefix(trimmedPath, basePath+"/") {
				return strings.TrimPrefix(trimmedPath, basePath+"/"), nil
			}
		}
		return trimmedPath, nil
	}
	if strings.TrimSpace(s.endpoint) != "" {
		prefix := s.bucket + "/"
		if strings.HasPrefix(trimmedPath, prefix) {
			return strings.TrimPrefix(trimmedPath, prefix), nil
		}
		return "", fmt.Errorf("invalid object url")
	}

	return trimmedPath, nil
}
