package handlers

import (
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
	"io"
	"io/ioutil"
	"net/http"
	"net/http/httptest"
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

func testRequest(t *testing.T, ts *httptest.Server, method, path string, body io.Reader) (*http.Response, string) {
	req, err := http.NewRequest(method, ts.URL+path, body)
	require.NoError(t, err)

	client := &http.Client{
		CheckRedirect: func(req *http.Request, via []*http.Request) error {
			return http.ErrUseLastResponse
		},
	}

	resp, err := client.Do(req)
	require.NoError(t, err)

	respBody, err := ioutil.ReadAll(resp.Body)
	require.NoError(t, err)

	defer resp.Body.Close()

	return resp, string(respBody)
}
