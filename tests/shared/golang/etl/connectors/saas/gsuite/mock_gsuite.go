package gsuite_utility

import (
	"errors"
	"net/http"
)

type MockGSuiteFn func() (*http.Response, error)

type MockGSuiteClient struct {
	DirectoryUsersList MockGSuiteFn
}

func (c *MockGSuiteClient) Do(req *http.Request) (*http.Response, error) {
	if req.URL.Path == "/admin/directory/v1/users" {
		return c.DirectoryUsersList()
	}
	return nil, errors.New("Invalid path.")
}
