package handler

import (
	"encoding/json"
	"errors"
	"io"
	"net/http"
)

const defaultJSONBodyLimit = 1 << 20

func decodeJSONBody(w http.ResponseWriter, r *http.Request, dst any, maxBytes int64) error {
	if maxBytes <= 0 {
		maxBytes = defaultJSONBodyLimit
	}
	if r == nil || r.Body == nil {
		return io.EOF
	}

	r.Body = http.MaxBytesReader(w, r.Body, maxBytes)

	decoder := json.NewDecoder(r.Body)
	decoder.DisallowUnknownFields()

	if err := decoder.Decode(dst); err != nil {
		return err
	}

	if err := decoder.Decode(&struct{}{}); err != nil {
		if errors.Is(err, io.EOF) {
			return nil
		}
		return err
	}

	return io.ErrUnexpectedEOF
}
