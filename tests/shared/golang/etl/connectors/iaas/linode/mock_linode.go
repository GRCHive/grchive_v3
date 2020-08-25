package linode_utility

import (
	"errors"
	"net/http"
	"strings"
)

type MockLinodeFn func() (*http.Response, error)

type MockLinodeClient struct {
	AccountUsers MockLinodeFn
	UserGrants   map[string]MockLinodeFn
}

func (c *MockLinodeClient) Do(req *http.Request) (*http.Response, error) {
	if strings.HasPrefix(req.URL.Path, "/v4/account/users") {
		if strings.HasSuffix(req.URL.Path, "/grants") {
			splitData := strings.Split(req.URL.Path, "/")
			username := splitData[len(splitData)-2]
			return c.UserGrants[username]()
		} else {
			return c.AccountUsers()
		}
	}
	return nil, errors.New("Invalid path.")
}
