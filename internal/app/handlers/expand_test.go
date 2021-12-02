package handlers

import (
	"errors"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/thorgnir-go-study/go-musthave-shortener/internal/app/storage"
	"github.com/thorgnir-go-study/go-musthave-shortener/internal/app/storage/mocks"
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
		storage *mocks.URLStorageMock
		want    want
	}{
		{
			name: "should expand google",
			request: request{
				url:    "/shortGoogle",
				method: http.MethodGet,
			},
			storage: func() *mocks.URLStorageMock {
				urlStorage := new(mocks.URLStorageMock)
				urlStorage.On("Load", "shortGoogle").Return("http://google.com", nil).Once()
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
			storage: func() *mocks.URLStorageMock {
				urlStorage := new(mocks.URLStorageMock)
				urlStorage.On("Load", "nonexistentId").Return("", storage.ErrURLNotFound).Once()
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
			storage: func() *mocks.URLStorageMock {
				urlStorage := new(mocks.URLStorageMock)
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
				st = new(mocks.URLStorageMock)
			}
			r := NewRouter(st)
			ts := httptest.NewServer(r)
			defer ts.Close()

			res := testRequest(t, ts, tt.request.method, tt.request.url, nil)
			// statictest иначе ругается
			defer func() {
				err := res.Body.Close()
				require.NoError(t, err)
			}()

			assert.Equal(t, tt.want.statusCode, res.StatusCode)
			if res.StatusCode == http.StatusTemporaryRedirect {
				assert.Equal(t, tt.want.contentType, res.Header.Get("Content-Type"))
				assert.Equal(t, tt.want.locationHeader, res.Header.Get("Location"))
			}
			st.AssertExpectations(t)
		})
	}

}
