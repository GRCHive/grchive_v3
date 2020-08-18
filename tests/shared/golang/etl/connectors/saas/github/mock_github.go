package github_utility

import (
	"errors"
	"net/http"
)

type MockGithubFn func() (*http.Response, error)

type MockGithubClient struct {
	GraphQL MockGithubFn
}

func (c *MockGithubClient) Do(req *http.Request) (*http.Response, error) {
	if req.URL.Path == "/graphql" {
		return c.GraphQL()
	}
	return nil, errors.New("Invalid path.")
}
