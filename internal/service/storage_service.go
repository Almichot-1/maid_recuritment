package service

import (
	"context"
	"errors"
	"fmt"
	"io"
	"mime"
	"net/url"
	"path"
	"strings"
	"time"

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
	presignClient *s3.PresignClient
	bucket        string
	region        string
	endpoint      string
	publicBaseURL string
}

var allowedContentTypes = map[string]struct{}{
	"image/jpeg":      {},
	"image/png":       {},
	"image/webp":      {},
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
		presignClient: s3.NewPresignClient(client),
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

type ReadURLOptions struct {
	FileName    string
	ContentType string
	Inline      bool
	Expires     time.Duration
}

func (s *S3StorageService) Open(fileURL string) (io.ReadCloser, string, error) {
	key, err := s.extractObjectKey(fileURL)
	if err != nil {
		return nil, "", err
	}

	result, err := s.client.GetObject(context.Background(), &s3.GetObjectInput{
		Bucket: aws.String(s.bucket),
		Key:    aws.String(key),
	})
	if err != nil {
		return nil, "", fmt.Errorf("open object: %w", err)
	}

	contentType := strings.TrimSpace(aws.ToString(result.ContentType))
	if contentType == "" {
		contentType = mime.TypeByExtension(path.Ext(key))
	}
	if contentType == "" {
		contentType = "application/octet-stream"
	}

	return result.Body, contentType, nil
}

func (s *S3StorageService) SignedURL(fileURL string, options ReadURLOptions) (string, error) {
	key, err := s.extractObjectKey(fileURL)
	if err != nil {
		return "", err
	}

	ttl := options.Expires
	if ttl <= 0 {
		ttl = 10 * time.Minute
	}
	if ttl > time.Hour {
		ttl = time.Hour
	}

	input := &s3.GetObjectInput{
		Bucket: aws.String(s.bucket),
		Key:    aws.String(key),
	}
	if contentType := strings.TrimSpace(options.ContentType); contentType != "" {
		input.ResponseContentType = aws.String(contentType)
	}
	if fileName := sanitizeDownloadFileName(options.FileName, key); fileName != "" {
		dispositionType := "attachment"
		if options.Inline {
			dispositionType = "inline"
		}
		input.ResponseContentDisposition = aws.String(fmt.Sprintf(`%s; filename="%s"`, dispositionType, fileName))
	}

	result, err := s.presignClient.PresignGetObject(context.Background(), input, func(opts *s3.PresignOptions) {
		opts.Expires = ttl
	})
	if err != nil {
		return "", fmt.Errorf("presign object: %w", err)
	}

	return result.URL, nil
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
	if strings.HasPrefix(strings.TrimSpace(fileURL), "s3://") {
		trimmed := strings.TrimPrefix(strings.TrimSpace(fileURL), "s3://")
		prefix := s.bucket + "/"
		if strings.HasPrefix(trimmed, prefix) {
			return strings.TrimPrefix(trimmed, prefix), nil
		}
		return trimmed, nil
	}

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

func sanitizeDownloadFileName(fileName, objectKey string) string {
	name := strings.TrimSpace(fileName)
	if name == "" {
		name = path.Base(objectKey)
	}
	name = strings.Map(func(r rune) rune {
		switch r {
		case '"', '\\', '\r', '\n':
			return -1
		default:
			return r
		}
	}, name)
	return strings.TrimSpace(name)
}
