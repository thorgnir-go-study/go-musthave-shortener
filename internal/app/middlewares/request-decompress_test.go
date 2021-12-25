package middlewares

import (
	"bytes"
	"compress/gzip"
	"github.com/go-chi/chi/v5"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestRequestDecompress(t *testing.T) {
	r := chi.NewRouter()

	r.Use(GzipRequestDecompressor)

	r.Post("/testHandler", simpleHandler)

	ts := httptest.NewServer(r)
	defer ts.Close()

	tests := []struct {
		name        string
		requestBody string
		want        string
		compress    bool
	}{
		{
			name:        "compressed",
			requestBody: "test data",
			want:        "test data",
			compress:    true,
		},
		{
			name:        "plain",
			requestBody: "test data",
			want:        "test data",
			compress:    false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			requestBody, err := getRequestBody(tt.requestBody, tt.compress)
			require.NoError(t, err)
			req, err := http.NewRequest("POST", ts.URL+"/testHandler", bytes.NewReader(requestBody))
			require.NoError(t, err)
			if tt.compress {
				req.Header.Set("Content-Encoding", "gzip")
				req.Header.Set("Content-Type", "application/x-gzip")
			}

			resp, err := http.DefaultClient.Do(req)
			require.NoError(t, err)
			assert.Equal(t, 200, resp.StatusCode)
			respBody, err := io.ReadAll(resp.Body)
			defer resp.Body.Close()
			require.NoError(t, err)
			assert.Equal(t, tt.want, string(respBody))
		})
	}
}

func simpleHandler(w http.ResponseWriter, r *http.Request) {
	bodyContent, err := io.ReadAll(r.Body)
	defer r.Body.Close()
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	w.Header().Set("Content-Type", "text/plain")
	_, err = w.Write(bodyContent)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
}

func gzipCompressString(input string) ([]byte, error) {
	var b bytes.Buffer
	gz := gzip.NewWriter(&b)
	_, err := gz.Write([]byte(input))
	if err != nil {
		return nil, err
	}
	err = gz.Close()
	if err != nil {
		return nil, err
	}
	return b.Bytes(), nil
}

func getRequestBody(input string, compress bool) ([]byte, error) {
	if compress {
		return gzipCompressString(input)
	}
	return []byte(input), nil
}
