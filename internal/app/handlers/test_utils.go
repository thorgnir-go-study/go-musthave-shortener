package handlers

import (
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
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

func testRequest(t *testing.T, ts *httptest.Server, method, path string, body string) *http.Response {
	bodyReader := strings.NewReader(body)

	req, err := http.NewRequest(method, ts.URL+path, bodyReader)
	require.NoError(t, err)

	client := &http.Client{
		CheckRedirect: func(req *http.Request, via []*http.Request) error {
			return http.ErrUseLastResponse
		},
	}

	resp, err := client.Do(req)
	require.NoError(t, err)

	return resp
}
