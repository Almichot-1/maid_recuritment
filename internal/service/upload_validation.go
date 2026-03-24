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

	return io.MultiReader(bytes.NewReader(header), file), actualContentType, nil
}

func detectContentTypeFromBytes(header []byte) (string, error) {
	switch {
	case bytes.HasPrefix(header, []byte{0xFF, 0xD8, 0xFF}):
		return "image/jpeg", nil
	case bytes.HasPrefix(header, []byte{0x89, 'P', 'N', 'G', '\r', '\n', 0x1a, '\n'}):
		return "image/png", nil
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
	case "video/mp4":
		return ".mp4"
	case "application/pdf":
		return ".pdf"
	default:
		return ""
	}
}
