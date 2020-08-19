package gcloud_utility

import (
	"errors"
	"net/http"
	"strings"
)

type MockGCloudFn func() (*http.Response, error)

type MockGCloudClient struct {
	GetUserListing  MockGCloudFn
	RolePermissions map[string]MockGCloudFn
}

func (c *MockGCloudClient) Do(req *http.Request) (*http.Response, error) {
	if req.URL.Host == "cloudresourcemanager.googleapis.com" && req.URL.Path == "/v1/projects/test:getIamPolicy" {
		return c.GetUserListing()
	} else if req.URL.Host == "iam.googleapis.com" && strings.Contains(req.URL.Path, "roles") {
		role := strings.TrimPrefix(req.URL.Path, "/v1/")
		return c.RolePermissions[role]()
	}
	return nil, errors.New("Invalid path.")
}
