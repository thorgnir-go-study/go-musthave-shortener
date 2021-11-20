package handlers

import (
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"
)

func Test_ShortenURLHandler(t *testing.T) {
	type request struct {
		url    string
		method string
		body   string
	}
	type want struct {
		contentType string
		statusCode  int
		body        string
	}
	tests := []struct {
		name    string
		request request
		want    want
	}{
		{
			name: "should shorten google",
			request: request{
				url:    "/",
				method: http.MethodPost,
				body:   "http://google.com",
			},
			want: want{
				contentType: "text/plain; charset=utf-8",
				statusCode:  http.StatusCreated,
				body:        "http://localhost:8080/shortGoogle",
			},
		},
		{
			name: "should fail on empty body",
			request: request{
				url:    "/",
				method: http.MethodPost,
				body:   "",
			},
			want: want{
				statusCode: http.StatusBadRequest,
			},
		},
		{
			name: "should fail on relative url",
			request: request{
				url:    "/",
				method: http.MethodPost,
				body:   "/somerelativeurl",
			},
			want: want{
				statusCode: http.StatusBadRequest,
			},
		},
		{
			name: "should fail on invalid url",
			request: request{
				url:    "/",
				method: http.MethodPost,
				body:   "some text",
			},
			want: want{
				statusCode: http.StatusBadRequest,
			},
		},
	}
	urlStorage := new(URLStorageMock)
	urlStorage.On("Store", "http://google.com").Return("shortGoogle", nil).Once()
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			r := NewRouter(urlStorage)
			ts := httptest.NewServer(r)
			defer ts.Close()
			res := testRequest(t, ts, tt.request.method, tt.request.url, tt.request.body)

			assert.Equal(t, tt.want.statusCode, res.StatusCode)
			if res.StatusCode == http.StatusCreated {
				assert.Equal(t, tt.want.contentType, res.Header.Get("Content-Type"))
				defer res.Body.Close()
				body, err := io.ReadAll(res.Body)
				require.NoError(t, err)
				assert.Equal(t, tt.want.body, string(body))
			}
		})
	}

	urlStorage.AssertExpectations(t)
}
