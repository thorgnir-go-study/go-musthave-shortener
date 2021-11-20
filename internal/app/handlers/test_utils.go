package handlers

import "github.com/stretchr/testify/mock"

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
