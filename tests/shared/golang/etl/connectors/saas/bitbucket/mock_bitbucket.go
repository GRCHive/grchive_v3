package bitbucket_utility

import (
	"errors"
	"net/http"
)

type MockBitbucketFn func() (*http.Response, error)

type MockBitbucketClient struct {
	WorkspaceMembersAll MockBitbucketFn
}

func (c *MockBitbucketClient) Do(req *http.Request) (*http.Response, error) {
	if req.URL.Path == "/2.0/workspaces/test/permissions" {
		return c.WorkspaceMembersAll()
	}
	return nil, errors.New("Invalid path.")
}
