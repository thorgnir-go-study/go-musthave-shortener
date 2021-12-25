package middlewares

import (
	"compress/gzip"
	"io"
	"net/http"
	"strings"
)

type gzipDecompressResponseReader struct {
	*gzip.Reader
	io.Closer
}

func (gz gzipDecompressResponseReader) Close() error {
	return gz.Closer.Close()
}

func GzipRequestDecompressor(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if strings.Contains(r.Header.Get("Content-Encoding"), "gzip") {
			gzr, err := gzip.NewReader(r.Body)
			if err != nil {
				http.Error(w, err.Error(), http.StatusInternalServerError)
			}
			r.Body = gzipDecompressResponseReader{gzr, r.Body}
		}
		next.ServeHTTP(w, r)
	})
}
