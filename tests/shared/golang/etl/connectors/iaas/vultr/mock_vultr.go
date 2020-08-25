package vultr_utility

import (
	"errors"
	"net/http"
)

type MockVultrFn func() (*http.Response, error)

type MockVultrClient struct {
	GetUsers       MockVultrFn
	GetAccountInfo MockVultrFn
}

func (c *MockVultrClient) Do(req *http.Request) (*http.Response, error) {
	if req.URL.Path == "/v2/users" {
		return c.GetUsers()
	} else if req.URL.Path == "/v2/account" {
		return c.GetAccountInfo()
	}
	return nil, errors.New("Invalid path.")
}
