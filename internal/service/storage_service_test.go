package service

import (
	"context"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/aws/aws-sdk-go-v2/aws"
	awsconfig "github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/credentials"
	"github.com/aws/aws-sdk-go-v2/service/s3"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"maid-recruitment-tracking/internal/config"
)

func TestNewS3StorageService_Validation(t *testing.T) {
	_, err := NewS3StorageService(nil)
	require.Error(t, err)

	_, err = NewS3StorageService(&config.Config{})
	require.Error(t, err)

	_, err = NewS3StorageService(&config.Config{S3Bucket: "b", AWSRegion: "r"})
	require.Error(t, err)
}

func TestS3StorageService_PublicURLAndExtractKey(t *testing.T) {
	service := &S3StorageService{bucket: "bucket1", region: "us-east-1", endpoint: "https://minio.local"}
	assert.Equal(t, "https://minio.local/bucket1/documents/file.pdf", service.publicURL("documents/file.pdf"))

	key, err := service.extractObjectKey("https://minio.local/bucket1/documents/file.pdf")
	require.NoError(t, err)
	assert.Equal(t, "documents/file.pdf", key)

	_, err = service.extractObjectKey("https://minio.local/another/documents/file.pdf")
	require.Error(t, err)

	service.endpoint = ""
	key, err = service.extractObjectKey("https://bucket1.s3.us-east-1.amazonaws.com/documents/file.pdf")
	require.NoError(t, err)
	assert.Equal(t, "documents/file.pdf", key)
}

func TestS3StorageService_UploadValidation(t *testing.T) {
	service := &S3StorageService{}
	_, err := service.Upload(strings.NewReader("x"), "file.bin", "application/octet-stream")
	require.ErrorIs(t, err, ErrUnsupportedContentType)

	_, err = service.Upload(nil, "file.jpg", "image/jpeg")
	require.Error(t, err)
}

func TestS3StorageService_DeleteValidation(t *testing.T) {
	service := &S3StorageService{bucket: "bucket1", endpoint: "https://minio.local"}
	err := service.Delete("::::")
	require.Error(t, err)
}

func newTestS3Client(t *testing.T, endpoint string) *s3.Client {
	t.Helper()
	cfg, err := awsconfig.LoadDefaultConfig(
		context.Background(),
		awsconfig.WithRegion("us-east-1"),
		awsconfig.WithCredentialsProvider(credentials.NewStaticCredentialsProvider("key", "secret", "")),
	)
	require.NoError(t, err)

	return s3.NewFromConfig(cfg, func(o *s3.Options) {
		o.BaseEndpoint = aws.String(endpoint)
		o.UsePathStyle = true
	})
}

func TestS3StorageService_UploadAndDelete_WithStubServer(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))
	defer server.Close()

	service := &S3StorageService{
		client:   newTestS3Client(t, server.URL),
		bucket:   "bucket1",
		region:   "us-east-1",
		endpoint: server.URL,
	}

	url, err := service.Upload(strings.NewReader("abc"), "a.pdf", "application/pdf")
	require.NoError(t, err)
	assert.Contains(t, url, "/bucket1/documents/")

	err = service.Delete(url)
	require.NoError(t, err)
}

func TestS3StorageService_UploadAndDelete_ServerFailure(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusInternalServerError)
		_, _ = io.WriteString(w, "fail")
	}))
	defer server.Close()

	service := &S3StorageService{
		client:   newTestS3Client(t, server.URL),
		bucket:   "bucket1",
		region:   "us-east-1",
		endpoint: server.URL,
	}

	_, err := service.Upload(strings.NewReader("abc"), "a.pdf", "application/pdf")
	require.Error(t, err)

	err = service.Delete(server.URL + "/bucket1/documents/a.pdf")
	require.Error(t, err)
}