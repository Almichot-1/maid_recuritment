package service

import (
	"bytes"
	"io"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestValidateAndBufferUploadReturnsSeekableBufferedReader(t *testing.T) {
	content := validPNGBytes()

	buffered, contentType, err := validateAndBufferUpload(bytes.NewReader(content), "candidate-photo.png")
	require.NoError(t, err)
	require.Equal(t, "image/png", contentType)

	readSeeker, ok := buffered.(io.ReadSeeker)
	require.True(t, ok, "buffered upload reader should support seeking for object-storage SDKs")

	result, err := io.ReadAll(readSeeker)
	require.NoError(t, err)
	require.Equal(t, content, result)

	_, err = readSeeker.Seek(0, io.SeekStart)
	require.NoError(t, err)

	replayed, err := io.ReadAll(readSeeker)
	require.NoError(t, err)
	require.Equal(t, content, replayed)
}
