package service

import (
	"bytes"
	"fmt"
	"io"
	"strings"
)

const fileSignatureBytes = 512

func validateAndBufferUpload(file io.Reader, fileName string) (io.Reader, string, error) {
	if file == nil {
		return nil, "", fmt.Errorf("file is required")
	}
	if strings.TrimSpace(fileName) == "" {
		return nil, "", fmt.Errorf("file name is required")
	}

	expectedContentType, err := detectContentTypeFromFileName(fileName)
	if err != nil {
		return nil, "", err
	}

	header := make([]byte, fileSignatureBytes)
	readBytes, readErr := io.ReadFull(file, header)
	switch readErr {
	case nil, io.EOF, io.ErrUnexpectedEOF:
	default:
		return nil, "", fmt.Errorf("read upload header: %w", readErr)
	}
	header = header[:readBytes]

	actualContentType, err := detectContentTypeFromBytes(header)
	if err != nil {
		return nil, "", err
	}
	if actualContentType != expectedContentType {
		return nil, "", ErrInvalidFileType
	}

	remaining, err := io.ReadAll(file)
	if err != nil {
		return nil, "", fmt.Errorf("buffer upload body: %w", err)
	}

	buffered := make([]byte, 0, len(header)+len(remaining))
	buffered = append(buffered, header...)
	buffered = append(buffered, remaining...)

	return bytes.NewReader(buffered), actualContentType, nil
}

func ValidateAndBufferUploadForProfile(file io.Reader, fileName string) (io.Reader, string, error) {
	buffered, contentType, err := validateAndBufferUpload(file, fileName)
	if err != nil {
		return nil, "", err
	}
	switch contentType {
	case "image/jpeg", "image/png", "image/webp":
		return buffered, contentType, nil
	default:
		return nil, "", ErrInvalidFileType
	}
}

func detectContentTypeFromBytes(header []byte) (string, error) {
	switch {
	case bytes.HasPrefix(header, []byte{0xFF, 0xD8, 0xFF}):
		return "image/jpeg", nil
	case bytes.HasPrefix(header, []byte{0x89, 'P', 'N', 'G', '\r', '\n', 0x1a, '\n'}):
		return "image/png", nil
	case len(header) >= 12 && string(header[0:4]) == "RIFF" && string(header[8:12]) == "WEBP":
		return "image/webp", nil
	case bytes.HasPrefix(header, []byte("%PDF-")):
		return "application/pdf", nil
	case len(header) >= 12 && string(header[4:8]) == "ftyp":
		return "video/mp4", nil
	default:
		return "", ErrInvalidFileType
	}
}

func extensionForContentType(contentType string) string {
	switch strings.TrimSpace(contentType) {
	case "image/jpeg":
		return ".jpg"
	case "image/png":
		return ".png"
	case "image/webp":
		return ".webp"
	case "video/mp4":
		return ".mp4"
	case "application/pdf":
		return ".pdf"
	default:
		return ""
	}
}
