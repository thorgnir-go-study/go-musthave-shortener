package handlers

import (
	"errors"
	"github.com/stretchr/testify/assert"
	"github.com/thorgnir-go-study/go-musthave-shortener/internal/app/config"
	shortenerMocks "github.com/thorgnir-go-study/go-musthave-shortener/internal/app/shortener/mocks"
	"github.com/thorgnir-go-study/go-musthave-shortener/internal/app/storage"
	storageMocks "github.com/thorgnir-go-study/go-musthave-shortener/internal/app/storage/mocks"
	"net/http"
	"net/http/httptest"
	"testing"
)

func Test_ExpandURLHandler(t *testing.T) {
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
		storage *storageMocks.URLStorager
		want    want
	}{
		{
			name: "should expand google",
			request: request{
				url:    "/shortGoogle",
				method: http.MethodGet,
			},
			storage: func() *storageMocks.URLStorager {
				urlStorage := new(storageMocks.URLStorager)
				urlStorage.On("Load", "shortGoogle").Return(storage.URLEntity{OriginalURL: "http://google.com"}, nil).Once()
				return urlStorage
			}(),
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
				statusCode: http.StatusMethodNotAllowed,
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
			storage: func() *storageMocks.URLStorager {
				urlStorage := new(storageMocks.URLStorager)
				urlStorage.On("Load", "nonexistentId").Return(storage.URLEntity{}, storage.ErrURLNotFound).Once()
				return urlStorage
			}(),
			want: want{
				statusCode: http.StatusNotFound,
			},
		},
		{
			name: "should respond 500 on storage load error",
			request: request{
				url:    "/short",
				method: http.MethodGet,
			},
			storage: func() *storageMocks.URLStorager {
				urlStorage := new(storageMocks.URLStorager)
				urlStorage.On("Load", "short").Return("", errors.New("Some error")).Once()
				return urlStorage
			}(),
			want: want{
				statusCode: http.StatusInternalServerError,
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			st := tt.storage
			if st == nil {
				st = new(storageMocks.URLStorager)
			}
			cfg := config.Config{}
			idGenerator := new(shortenerMocks.URLIDGenerator)
			service := NewService(st, idGenerator, cfg)
			r := NewRouter(service)
			ts := httptest.NewServer(r)
			defer ts.Close()

			res := testRequest(t, ts, tt.request.method, tt.request.url, nil)
			defer res.Body.Close()

			assert.Equal(t, tt.want.statusCode, res.StatusCode)
			if res.StatusCode == http.StatusTemporaryRedirect {
				assert.Equal(t, tt.want.contentType, res.Header.Get("Content-Type"))
				assert.Equal(t, tt.want.locationHeader, res.Header.Get("Location"))
			}
			st.AssertExpectations(t)
		})
	}

}
