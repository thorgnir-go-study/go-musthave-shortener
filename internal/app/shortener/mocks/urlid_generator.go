// Code generated by mockery v2.9.4. DO NOT EDIT.

package mocks

import mock "github.com/stretchr/testify/mock"

// URLIDGenerator is an autogenerated mock type for the URLIDGenerator type
type URLIDGenerator struct {
	mock.Mock
}

// GenerateURLID provides a mock function with given fields: originalURL
func (_m *URLIDGenerator) GenerateURLID(originalURL string) string {
	ret := _m.Called(originalURL)

	var r0 string
	if rf, ok := ret.Get(0).(func(string) string); ok {
		r0 = rf(originalURL)
	} else {
		r0 = ret.Get(0).(string)
	}

	return r0
}
