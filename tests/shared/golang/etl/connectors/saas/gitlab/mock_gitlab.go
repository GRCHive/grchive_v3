package gitlab_utility

import (
	"errors"
	"net/http"
)

type MockGitlabFn func() (*http.Response, error)

type MockGitlabClient struct {
	GroupMembersAll MockGitlabFn
}

func (c *MockGitlabClient) Do(req *http.Request) (*http.Response, error) {
	if req.URL.Path == "/api/v4/groups/test/members/all" {
		return c.GroupMembersAll()
	}
	return nil, errors.New("Invalid path.")
}
