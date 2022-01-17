package handlers

import (
	"errors"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
	"github.com/thorgnir-go-study/go-musthave-shortener/internal/app/config"
	storageMocks "github.com/thorgnir-go-study/go-musthave-shortener/internal/app/repository/mocks"
	shortenerMocks "github.com/thorgnir-go-study/go-musthave-shortener/internal/app/shortener/mocks"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func Test_JSONShortenURLHandler(t *testing.T) {
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
		name        string
		request     request
		want        want
		storage     *storageMocks.URLStorager
		idGenerator *shortenerMocks.URLIDGenerator
	}{
		{
			name: "should shorten google",
			request: request{
				url:    "/api/shorten",
				method: http.MethodPost,
				body:   `{"url": "http://google.com"}`,
			},
			want: want{
				contentType: "application/json; charset=utf-8",
				statusCode:  http.StatusCreated,
				body:        `{"result":"http://localhost:8080/shortGoogle"}`,
			},
			storage: func() *storageMocks.URLStorager {
				urlStorage := new(storageMocks.URLStorager)
				urlStorage.On("Store", mock.Anything, mock.Anything).Return(nil).Once()
				return urlStorage
			}(),
			idGenerator: func() *shortenerMocks.URLIDGenerator {
				gen := new(shortenerMocks.URLIDGenerator)
				gen.On("GenerateURLID", "http://google.com").Return("shortGoogle").Once()
				return gen
			}(),
		},
		{
			name: "should fail on empty body",
			request: request{
				url:    "/api/shorten",
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
				url:    "/api/shorten",
				method: http.MethodPost,
				body:   `{"url": "/somerelativeurl"}`,
			},
			want: want{
				statusCode: http.StatusBadRequest,
			},
		},
		{
			name: "should fail on invalid url",
			request: request{
				url:    "/api/shorten",
				method: http.MethodPost,
				body:   `{"url": "some text"}`,
			},
			want: want{
				statusCode: http.StatusBadRequest,
			},
		},
		{
			name: "should respond 500 on url repository error",
			request: request{
				url:    "/api/shorten",
				method: http.MethodPost,
				body:   `{"url": "http://google.com"}`,
			},
			want: want{
				statusCode: http.StatusInternalServerError,
			},
			storage: func() *storageMocks.URLStorager {
				urlStorage := new(storageMocks.URLStorager)
				urlStorage.On("Store", mock.Anything, mock.Anything).Return("", errors.New("some error")).Once()
				return urlStorage
			}(),
			idGenerator: func() *shortenerMocks.URLIDGenerator {
				gen := new(shortenerMocks.URLIDGenerator)
				gen.On("GenerateURLID", "http://google.com").Return("shortGoogle").Once()
				return gen
			}(),
		},
	}
	baseURL := "http://localhost:8080"

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			st := tt.storage
			if st == nil {
				st = new(storageMocks.URLStorager)
			}
			gen := tt.idGenerator
			if gen == nil {
				gen = new(shortenerMocks.URLIDGenerator)
			}
			cfg := config.Config{
				BaseURL: baseURL,
			}

			service := NewService(st, gen, cfg)
			r := NewRouter(service)
			ts := httptest.NewServer(r)
			defer ts.Close()
			res := testRequest(t, ts, tt.request.method, tt.request.url, strings.NewReader(tt.request.body))

			assert.Equal(t, tt.want.statusCode, res.StatusCode)
			if res.StatusCode == http.StatusCreated {
				assert.Equal(t, tt.want.contentType, res.Header.Get("Content-Type"))
				defer res.Body.Close()
				body, err := io.ReadAll(res.Body)
				require.NoError(t, err)
				assert.JSONEq(t, tt.want.body, string(body))
			}
			st.AssertExpectations(t)
		})
	}

}
