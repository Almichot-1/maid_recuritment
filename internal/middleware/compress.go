package middleware

import (
	"compress/gzip"
	"io"
	"net/http"
	"strings"
	"sync"
)

type gzipResponseWriter struct {
	io.Writer
	http.ResponseWriter
	statusCode int
	written    bool
}

func (w *gzipResponseWriter) WriteHeader(statusCode int) {
	w.statusCode = statusCode
	if statusCode < 300 || statusCode == 304 {
		w.ResponseWriter.Header().Del("Content-Length")
	}
	w.ResponseWriter.WriteHeader(statusCode)
}

func (w *gzipResponseWriter) Write(b []byte) (int, error) {
	if w.written {
		return w.Writer.Write(b)
	}
	w.written = true
	if w.statusCode == 0 {
		w.statusCode = http.StatusOK
	}
	return w.Writer.Write(b)
}

func (w *gzipResponseWriter) Flush() {
	if f, ok := w.ResponseWriter.(http.Flusher); ok {
		f.Flush()
	}
}

var gzipWriterPool = &sync.Pool{
	New: func() interface{} {
		w, _ := gzip.NewWriterLevel(io.Discard, gzip.BestSpeed)
		return w
	},
}

func Compress(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if strings.EqualFold(r.Header.Get("Upgrade"), "websocket") {
			next.ServeHTTP(w, r)
			return
		}

		if !strings.Contains(r.Header.Get("Accept-Encoding"), "gzip") {
			next.ServeHTTP(w, r)
			return
		}

		contentType := w.Header().Get("Content-Type")
		if contentType != "" && !shouldCompress(contentType) {
			next.ServeHTTP(w, r)
			return
		}

		if r.Method == "HEAD" || r.Method == "CONNECT" {
			next.ServeHTTP(w, r)
			return
		}

		gz := gzipWriterPool.Get().(*gzip.Writer)
		defer gzipWriterPool.Put(gz)
		gz.Reset(w)

		defer gz.Close()

		w.Header().Set("Content-Encoding", "gzip")
		w.Header().Del("Content-Length")

		gw := &gzipResponseWriter{Writer: gz, ResponseWriter: w}
		next.ServeHTTP(gw, r)
	})
}

func shouldCompress(contentType string) bool {
	compressible := []string{
		"text/",
		"application/json",
		"application/javascript",
		"application/x-javascript",
		"application/xml",
		"application/xhtml+xml",
		"application/ld+json",
		"application/manifest+json",
		"application/vnd.api+json",
		"image/svg+xml",
		"font/",
		"application/font-woff",
		"application/font-woff2",
		"application/vnd.ms-fontobject",
	}
	for _, ct := range compressible {
		if strings.HasPrefix(contentType, ct) {
			return true
		}
	}
	return false
}
