package handler

import (
	"fmt"
	"io"
	"mime"
	"net/http"
	"path/filepath"
	"strings"
	"time"

	"maid-recruitment-tracking/internal/service"
)

type secureDocumentStorage interface {
	Open(fileURL string) (io.ReadCloser, string, error)
	SignedURL(fileURL string, options service.ReadURLOptions) (string, error)
}

func absoluteURL(r *http.Request, targetPath string) string {
	if r == nil {
		return strings.TrimSpace(targetPath)
	}

	scheme := "http"
	if strings.EqualFold(strings.TrimSpace(r.Header.Get("X-Forwarded-Proto")), "https") || r.TLS != nil {
		scheme = "https"
	}

	return fmt.Sprintf("%s://%s%s", scheme, strings.TrimSpace(r.Host), strings.TrimSpace(targetPath))
}

func buildSignedDocumentURL(storage secureDocumentStorage, fileURL, fileName, contentType string, inline bool) string {
	if storage == nil || strings.TrimSpace(fileURL) == "" {
		return strings.TrimSpace(fileURL)
	}

	signedURL, err := storage.SignedURL(fileURL, service.ReadURLOptions{
		FileName:    fileName,
		ContentType: contentType,
		Inline:      inline,
		Expires:     15 * time.Minute,
	})
	if err != nil {
		return strings.TrimSpace(fileURL)
	}
	return signedURL
}

func contentTypeFromFileName(fileName string) string {
	if strings.TrimSpace(fileName) == "" {
		return ""
	}
	return mime.TypeByExtension(strings.ToLower(filepath.Ext(fileName)))
}
