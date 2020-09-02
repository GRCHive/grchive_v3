package azure

import (
	"fmt"
	"github.com/onsi/gomega"
	"gitlab.com/grchive/grchive-v3/shared/etl/types"
	"gitlab.com/grchive/grchive-v3/shared/test_utility"
	"gitlab.com/grchive/grchive-v3/tests/shared/etl/connectors/iaas/azure_utility"
	"net/http"
	"testing"
	"time"
)

var refTime1 = time.Date(2010, 10, 12, 15, 34, 33, 0, time.UTC)
var refTime2 = time.Date(2010, 12, 10, 3, 5, 6, 1, time.UTC)

func createGraphClient() *azure_utility.MockAzureGraphClient {
	return &azure_utility.MockAzureGraphClient{
		UsersList: func() (*http.Response, error) {
			return test_utility.WrapHttpResponse(fmt.Sprintf(`
{"@odata.context":"https://graph.microsoft.com/v1.0/$metadata#users(displayName,userPrincipalName,mail,otherMails,createdDateTime,id)","value":[{"displayName":"Michael Bao","userPrincipalName":"mike_grchive.com#EXT#@mikegrchive.onmicrosoft.com","mail":null,"otherMails":["mike@grchive.com"],"createdDateTime":"%s","id":"1e7ef588-4893-486c-a879-00af5d017734"}]}
`, refTime1.Format(time.RFC3339))), nil
		},
		DirectoryRoles: func() (*http.Response, error) {
			return test_utility.WrapHttpResponse(`
{"@odata.context":"https://graph.microsoft.com/v1.0/$metadata#directoryRoles","value":[{"id":"01b64ea3-a49a-4254-9d71-69094919d3a3","deletedDateTime":null,"description":"Can manage all aspects of the Exchange product.","displayName":"Exchange Service Administrator","roleTemplateId":"29232cdf-9323-42fd-ade2-1d097af3e4de"}]}
`), nil
		},
		DirectoryRoleMembers: map[string]azure_utility.MockAzureFn{
			"01b64ea3-a49a-4254-9d71-69094919d3a3": func() (*http.Response, error) {
				return test_utility.WrapHttpResponse(`
{"@odata.context":"https://graph.microsoft.com/v1.0/$metadata#directoryObjects","value":[{"@odata.type":"#microsoft.graph.user","id":"1e7ef588-4893-486c-a879-00af5d017734"}]}
`), nil
			},
		},
	}
}

func createManagementClient() *azure_utility.MockAzureManagementClient {
	return &azure_utility.MockAzureManagementClient{
		UserAppRoleAssignments: map[string]azure_utility.MockAzureFn{
			"1e7ef588-4893-486c-a879-00af5d017734": func() (*http.Response, error) {
				return test_utility.WrapHttpResponse(`
{"value":[{"properties":{"roleDefinitionId":"/subscriptions/38b08a9b-c63b-4848-b4ae-4c83b6f7f855/providers/Microsoft.Authorization/roleDefinitions/8e3af657-a8ff-443c-a75c-2fe8c4bcb635","principalId":"1e7ef588-4893-486c-a879-00af5d017734","scope":"/subscriptions/38b08a9b-c63b-4848-b4ae-4c83b6f7f855","createdOn":"2020-08-24T14:57:46.3292144Z", "updatedOn":"2020-08-24T14:57:46.3292144Z","createdBy":"","updatedBy":""},"id":"/subscriptions/38b08a9b-c63b-4848-b4ae-4c83b6f7f855/providers/Microsoft.Authorization/roleAssignments/7b31b142-dc79-4ce7-9b7f-55f6ff6f78ec","type":"Microsoft.Authorization/roleAssignments","name":"7b31b142-dc79-4ce7-9b7f-55f6ff6f78ec"}]}
`), nil
			},
		},
		RoleDefinition: map[string]azure_utility.MockAzureFn{
			"8e3af657-a8ff-443c-a75c-2fe8c4bcb635": func() (*http.Response, error) {
				return test_utility.WrapHttpResponse(`
{"properties":{"roleName":"Owner","type":"BuiltInRole","description"
:"Grants full access to manage all resources, including the ability to assign roles in Azure RBAC.","assignableScopes":["/"],"permissions":[{"actions":["*"],"notActions":[]}],"createdOn":"2015-02-02T21:55:09.8806423Z","updatedOn":"2020-08-14T20:13:58.4137852Z","createdBy":null,"updatedBy":null},"id":"/subscriptions/38b08a9b-c63b-4848-b4ae-4c83b6f7f855/providers/Microsoft.Authorization/roleDefinitions/8e3af657-a8ff-443c-a75c-2fe8c4bcb635","type":"Microsoft.Authorization/roleDefinitions","name":"8e3af657-a8ff-443c-a75c-2fe8c4bcb635"}
`), nil
			},
		},
	}
}

func createConnector(g *gomega.GomegaWithT) *EtlAzureConnector {
	conn, err := CreateAzureConnector(&EtlAzureOptions{
		GraphClient:      createGraphClient(),
		ManagementClient: createManagementClient(),
		SubscriptionId:   "test",
	})
	g.Expect(err).To(gomega.BeNil())
	return conn
}

func TestGetAllUsers(t *testing.T) {
	g := gomega.NewGomegaWithT(t)
	conn := createConnector(g)

	refUsers := []azureUser{
		azureUser{
			Id:                "1e7ef588-4893-486c-a879-00af5d017734",
			DisplayName:       "Michael Bao",
			UserPrincipalName: "mike_grchive.com#EXT#@mikegrchive.onmicrosoft.com",
			Mail:              "",
			OtherMails:        []string{"mike@grchive.com"},
			CreatedDateTime:   refTime1,
		},
	}

	users, source, err := conn.users.getAllUsers()
	g.Expect(err).To(gomega.BeNil())
	g.Expect(len(source.Commands)).To(gomega.Equal(1))
	g.Expect(users).To(gomega.Equal(refUsers))
}

func TestGetUserAppRoleAssignments(t *testing.T) {
	g := gomega.NewGomegaWithT(t)
	conn := createConnector(g)

	testUser := azureUser{
		Id:                "1e7ef588-4893-486c-a879-00af5d017734",
		DisplayName:       "Michael Bao",
		UserPrincipalName: "mike_grchive.com#EXT#@mikegrchive.onmicrosoft.com",
		Mail:              "",
		OtherMails:        []string{"mike@grchive.com"},
		CreatedDateTime:   refTime1,
	}

	refAssignments := []azureAppRoleAssignment{
		azureAppRoleAssignment{
			Id:   "/subscriptions/38b08a9b-c63b-4848-b4ae-4c83b6f7f855/providers/Microsoft.Authorization/roleAssignments/7b31b142-dc79-4ce7-9b7f-55f6ff6f78ec",
			Name: "7b31b142-dc79-4ce7-9b7f-55f6ff6f78ec",
			Properties: azureAppRoleAssignmentProperties{
				PrincipalId:      "1e7ef588-4893-486c-a879-00af5d017734",
				RoleDefinitionId: "/subscriptions/38b08a9b-c63b-4848-b4ae-4c83b6f7f855/providers/Microsoft.Authorization/roleDefinitions/8e3af657-a8ff-443c-a75c-2fe8c4bcb635",
				Scope:            "/subscriptions/38b08a9b-c63b-4848-b4ae-4c83b6f7f855",
			},
		},
	}

	assignments, source, err := conn.users.getUserAppRoleAssignments(testUser.Id)
	g.Expect(err).To(gomega.BeNil())
	g.Expect(len(source.Commands)).To(gomega.Equal(1))
	g.Expect(assignments).To(gomega.Equal(refAssignments))
}

func TestGetRoleDefinition(t *testing.T) {
	g := gomega.NewGomegaWithT(t)
	conn := createConnector(g)

	refDef := azureRoleDefinition{
		Properties: azureRoleDefinitionProperties{
			Name:   "Owner",
			Scopes: []string{"/"},
			Permissions: []azureRolePermission{
				azureRolePermission{
					Actions:    []string{"*"},
					NotActions: []string{},
				},
			},
		},
	}

	defId := "/subscriptions/38b08a9b-c63b-4848-b4ae-4c83b6f7f855/providers/Microsoft.Authorization/roleDefinitions/8e3af657-a8ff-443c-a75c-2fe8c4bcb635"

	def, source, err := conn.users.getRoleDefinition(defId)
	g.Expect(err).To(gomega.BeNil())
	g.Expect(len(source.Commands)).To(gomega.Equal(1))
	g.Expect(*def).To(gomega.Equal(refDef))
}

func TestToEtlRole(t *testing.T) {
	g := gomega.NewGomegaWithT(t)
	for _, test := range []struct {
		Definition *azureRoleDefinition
		Role       *types.EtlRole
	}{
		{
			Definition: &azureRoleDefinition{
				Properties: azureRoleDefinitionProperties{
					Name:   "Hello",
					Scopes: []string{"scope1", "scope2"},
					Permissions: []azureRolePermission{
						azureRolePermission{
							Actions: []string{"a1", "a2"},
						},
					},
				},
			},
			Role: &types.EtlRole{
				Name: "Hello",
				Permissions: map[string][]string{
					"scope1": []string{"a1", "a2"},
					"scope2": []string{"a1", "a2"},
				},
			},
		},
	} {
		cmp := test.Definition.toEtlRole()
		g.Expect(*cmp).To(gomega.Equal(*test.Role))
	}
}

func TestToEtlUser(t *testing.T) {
	g := gomega.NewGomegaWithT(t)
	for _, test := range []struct {
		Azure *azureUser
		User  *types.EtlUser
	}{
		{
			Azure: &azureUser{
				Id:                "asdf",
				DisplayName:       "1234",
				UserPrincipalName: "principal",
				Mail:              "",
				OtherMails:        []string{"mike@grchive.com", "test@grchive.com"},
				CreatedDateTime:   refTime1,
			},
			User: &types.EtlUser{
				Username:    "principal",
				Email:       "mike@grchive.com",
				FullName:    "1234",
				CreatedTime: &refTime1,
				Roles:       map[string]*types.EtlRole{},
			},
		},
		{
			Azure: &azureUser{
				Id:                "asdf",
				DisplayName:       "1234",
				UserPrincipalName: "principal",
				Mail:              "derek@grchive.com",
				OtherMails:        []string{"mike@grchive.com"},
				CreatedDateTime:   refTime2,
			},
			User: &types.EtlUser{
				Username:    "principal",
				Email:       "derek@grchive.com",
				FullName:    "1234",
				CreatedTime: &refTime2,
				Roles:       map[string]*types.EtlRole{},
			},
		},
	} {
		cmp := test.Azure.toEtlUser()
		g.Expect(*cmp).To(gomega.Equal(*test.User))
	}
}

func TestGetUserListing(t *testing.T) {
	g := gomega.NewGomegaWithT(t)
	conn := createConnector(g)

	itf, err := conn.GetUserInterface()
	g.Expect(err).To(gomega.BeNil())

	users, source, err := itf.GetUserListing()
	g.Expect(err).To(gomega.BeNil())

	g.Expect(err).To(gomega.BeNil())
	g.Expect(len(source.Commands)).To(gomega.Equal(5))

	refUsers := map[string]*types.EtlUser{
		"mike_grchive.com#EXT#@mikegrchive.onmicrosoft.com": &types.EtlUser{
			Username:    "mike_grchive.com#EXT#@mikegrchive.onmicrosoft.com",
			Email:       "mike@grchive.com",
			FullName:    "Michael Bao",
			CreatedTime: &refTime1,
			Roles: map[string]*types.EtlRole{
				"Owner": &types.EtlRole{
					Name: "Owner",
					Permissions: map[string][]string{
						"/": []string{"*"},
					},
				},
				"Exchange Service Administrator": &types.EtlRole{
					Name:        "Exchange Service Administrator",
					Permissions: map[string][]string{},
				},
			},
		},
	}
	test_utility.CompareUserListing(g, users, refUsers, test_utility.CompareUserListingOptions{})
}
