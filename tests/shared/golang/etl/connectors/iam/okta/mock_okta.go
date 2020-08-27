package okta_utility

import (
	"errors"
	"net/http"
	"strings"
)

type MockOktaFn func() (*http.Response, error)

type MockOktaClient struct {
	Users     MockOktaFn
	UserRoles map[string]MockOktaFn
}

func (c *MockOktaClient) Do(req *http.Request) (*http.Response, error) {
	if strings.HasPrefix(req.URL.Path, "/api/v1/users") {
		userPath := strings.TrimPrefix(req.URL.Path, "/api/v1/users")
		if userPath == "" || userPath == "/" {
			return c.Users()
		} else if strings.HasSuffix(userPath, "/roles") {
			userSplit := strings.Split(userPath, "/")
			user := userSplit[1]
			return c.UserRoles[user]()
		}
	}
	return nil, errors.New("Invalid path.")
}
