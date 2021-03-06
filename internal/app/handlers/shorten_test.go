package handlers

import (
	"errors"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
	"github.com/thorgnir-go-study/go-musthave-shortener/internal/app/config"
	repositoryMocks "github.com/thorgnir-go-study/go-musthave-shortener/internal/app/repository/mocks"
	shortenerMocks "github.com/thorgnir-go-study/go-musthave-shortener/internal/app/shortener/mocks"
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
		name        string
		request     request
		want        want
		storage     *repositoryMocks.URLRepository
		idGenerator *shortenerMocks.URLIDGenerator
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
			storage: func() *repositoryMocks.URLRepository {
				urlStorage := new(repositoryMocks.URLRepository)
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
			name: "should respond 500 on url repository error",
			request: request{
				url:    "/",
				method: http.MethodPost,
				body:   "http://google.com",
			},
			want: want{
				statusCode: http.StatusInternalServerError,
			},
			storage: func() *repositoryMocks.URLRepository {
				urlStorage := new(repositoryMocks.URLRepository)
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
				st = new(repositoryMocks.URLRepository)
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
				body, err := io.ReadAll(res.Body)
				defer res.Body.Close()
				require.NoError(t, err)
				assert.Equal(t, tt.want.body, string(body))
			}
			st.AssertExpectations(t)
		})
	}

}
