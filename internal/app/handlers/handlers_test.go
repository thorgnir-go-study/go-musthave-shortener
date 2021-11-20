package handlers

import (
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

type URLStorageMock struct {
	mock.Mock
}

func (m *URLStorageMock) Store(url string) (string, error) {
	args := m.Called(url)
	return args.String(0), args.Error(1)
}

func (m *URLStorageMock) Load(key string) (string, bool, error) {
	args := m.Called(key)
	return args.String(0), args.Bool(1), args.Error(2)
}

func Test_shortenURLHandler(t *testing.T) {
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
			request := httptest.NewRequest(tt.request.method, tt.request.url, strings.NewReader(tt.request.body))
			w := httptest.NewRecorder()
			h := RootHanlder(urlStorage)

			h.ServeHTTP(w, request)

			res := w.Result()

			assert.Equal(t, tt.want.statusCode, res.StatusCode)
			if res.StatusCode == http.StatusCreated {
				assert.Equal(t, tt.want.contentType, res.Header.Get("Content-Type"))
				defer res.Body.Close()
				resBody, err := io.ReadAll(res.Body)
				require.NoError(t, err)
				assert.Equal(t, tt.want.body, string(resBody))
			}
		})
	}

	urlStorage.AssertExpectations(t)
}

func Test_expandURLHandler(t *testing.T) {
	type request struct {
		url    string
		method string
	}
	type want struct {
		contentType    string
		statusCode     int
		locationHeader string
	}
	tests := []struct {
		name    string
		request request
		want    want
	}{
		{
			name: "should unshorten google",
			request: request{
				url:    "/shortGoogle",
				method: http.MethodGet,
			},
			want: want{
				contentType:    "text/plain; charset=utf-8",
				statusCode:     http.StatusTemporaryRedirect,
				locationHeader: "http://google.com",
			},
		},
		{
			name: "should fail on empty req",
			request: request{
				url:    "/",
				method: http.MethodGet,
			},
			want: want{
				statusCode: http.StatusBadRequest,
			},
		},
		{
			name: "should respond 404 on deep urls",
			request: request{
				url:    "/blabla/blabla",
				method: http.MethodGet,
			},
			want: want{
				statusCode: http.StatusNotFound,
			},
		},
		{
			name: "should respond 404 on unknown id",
			request: request{
				url:    "/nonexistentId",
				method: http.MethodGet,
			},
			want: want{
				statusCode: http.StatusNotFound,
			},
		},
	}
	urlStorage := new(URLStorageMock)
	urlStorage.On("Load", "shortGoogle").Return("http://google.com", true, nil).Once()
	urlStorage.On("Load", "nonexistentId").Return("", false, nil).Once()
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			request := httptest.NewRequest(tt.request.method, tt.request.url, nil)
			w := httptest.NewRecorder()
			h := RootHanlder(urlStorage)

			h.ServeHTTP(w, request)

			res := w.Result()

			assert.Equal(t, tt.want.statusCode, res.StatusCode)
			if res.StatusCode == http.StatusTemporaryRedirect {
				assert.Equal(t, tt.want.contentType, res.Header.Get("Content-Type"))
				assert.Equal(t, tt.want.locationHeader, res.Header.Get("Location"))
			}
		})
	}

	urlStorage.AssertExpectations(t)
}
