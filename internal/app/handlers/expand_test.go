package handlers

import (
	"github.com/stretchr/testify/assert"
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
		want    want
	}{
		{
			name: "should expand google",
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
			h := ExpandURLHandler(urlStorage)

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
