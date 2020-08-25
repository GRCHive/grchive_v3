package heroku_utility

import (
	"errors"
	"net/http"
)

type MockHerokuFn func() (*http.Response, error)

type MockHerokuClient struct {
	TeamMembers MockHerokuFn
}

func (c *MockHerokuClient) Do(req *http.Request) (*http.Response, error) {
	if req.URL.Path == "/teams/test/members" {
		return c.TeamMembers()
	}
	return nil, errors.New("Invalid path.")
}
