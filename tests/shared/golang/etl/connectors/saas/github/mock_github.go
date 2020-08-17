package github_utility

import (
	"errors"
	"net/http"
)

type MockGithubFn func() (*http.Response, error)

type MockGithubClient struct {
	OrgMembersList MockGithubFn
	OrgAdminsList  MockGithubFn
}

func (c *MockGithubClient) Do(req *http.Request) (*http.Response, error) {
	query := req.URL.Query()

	if req.URL.Path == "/orgs/test/members" {
		if query.Get("role") == "admin" {
			return c.OrgAdminsList()
		} else if query.Get("role") == "member" {
			return c.OrgMembersList()
		}
	}
	return nil, errors.New("Invalid path.")
}
