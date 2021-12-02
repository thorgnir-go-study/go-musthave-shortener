package handlers

import (
	"errors"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/thorgnir-go-study/go-musthave-shortener/internal/app/storage/mocks"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
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
		storage *mocks.URLStorage
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
			storage: func() *mocks.URLStorage {
				urlStorage := new(mocks.URLStorage)
				urlStorage.On("Store", "http://google.com").Return("shortGoogle", nil).Once()
				return urlStorage
			}(),
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
		{
			name: "should respond 500 on url storage error",
			request: request{
				url:    "/",
				method: http.MethodPost,
				body:   "http://google.com",
			},
			want: want{
				statusCode: http.StatusInternalServerError,
			},
			storage: func() *mocks.URLStorage {
				urlStorage := new(mocks.URLStorage)
				urlStorage.On("Store", "http://google.com").Return("", errors.New("some error")).Once()
				return urlStorage
			}(),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			st := tt.storage
			if st == nil {
				st = new(mocks.URLStorage)
			}

			r := NewRouter(st)
			ts := httptest.NewServer(r)
			defer ts.Close()
			res := testRequest(t, ts, tt.request.method, tt.request.url, strings.NewReader(tt.request.body))

			assert.Equal(t, tt.want.statusCode, res.StatusCode)
			if res.StatusCode == http.StatusCreated {
				assert.Equal(t, tt.want.contentType, res.Header.Get("Content-Type"))
				defer res.Body.Close()
				body, err := io.ReadAll(res.Body)
				require.NoError(t, err)
				assert.Equal(t, tt.want.body, string(body))
			}
			st.AssertExpectations(t)
		})
	}

}
