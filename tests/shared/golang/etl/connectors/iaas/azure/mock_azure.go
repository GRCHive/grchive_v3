package azure_utility

import (
	"errors"
	"net/http"
	"regexp"
	"strings"
)

type MockAzureFn func() (*http.Response, error)

type MockAzureGraphClient struct {
	UsersList MockAzureFn
}

func (c *MockAzureGraphClient) Do(req *http.Request) (*http.Response, error) {
	if req.URL.Path == "/v1.0/users" {
		return c.UsersList()
	}
	return nil, errors.New("Invalid path.")
}

type MockAzureManagementClient struct {
	UserAppRoleAssignments map[string]MockAzureFn
	RoleDefinition         map[string]MockAzureFn
}

func (c *MockAzureManagementClient) Do(req *http.Request) (*http.Response, error) {
	if req.URL.Path == "/subscriptions/test/providers/Microsoft.Authorization/roleAssignments" {
		assignedToRegex := regexp.MustCompile(`assignedTo\('(.*)'\)`)
		userId := assignedToRegex.FindStringSubmatch(req.URL.Query().Get("$filter"))[1]
		return c.UserAppRoleAssignments[userId]()
	} else if strings.Contains(req.URL.Path, "/roleDefinitions/") {
		defId := strings.Split(req.URL.Path, "/roleDefinitions/")[1]
		return c.RoleDefinition[defId]()
	}
	return nil, errors.New("Invalid path.")
}
