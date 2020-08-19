package cloudflare_utility

import (
	"errors"
	"net/http"
)

type MockCloudflareFn func() (*http.Response, error)

type MockCloudflareClient struct {
	AccountMembers MockCloudflareFn
}

func (c *MockCloudflareClient) Do(req *http.Request) (*http.Response, error) {
	if req.URL.Path == "/client/v4/accounts/test/members" {
		return c.AccountMembers()
	}
	return nil, errors.New("Invalid path.")
}
