package auth0_utility

import (
	"errors"
	"gitlab.com/grchive/grchive-v3/shared/test_utility"
	"net/http"
	"strings"
)

type MockAuth0Fn func() (*http.Response, error)

type MockAuth0Client struct {
	Users MockAuth0Fn
}

func (c *MockAuth0Client) Do(req *http.Request) (*http.Response, error) {
	if strings.HasPrefix(req.URL.Path, "/api/v2/users") {
		page := req.URL.Query().Get("page")
		if page == "0" {
			return c.Users()
		} else {
			return test_utility.WrapHttpResponse(`[]`), nil

		}
	}
	return nil, errors.New("Invalid path.")
}
