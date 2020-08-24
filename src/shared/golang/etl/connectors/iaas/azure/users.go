package azure

import (
	"fmt"
	"gitlab.com/grchive/grchive-v3/shared/etl/connectors"
	"gitlab.com/grchive/grchive-v3/shared/etl/types"
	"gitlab.com/grchive/grchive-v3/shared/utility/mt"
	"sync"
	"time"
)

type azureUser struct {
	Id                string    `json:"id"`
	DisplayName       string    `json:"displayName"`
	UserPrincipalName string    `json:"userPrincipalName"`
	Mail              string    `json:"mail"`
	OtherMails        []string  `json:"otherMails"`
	CreatedDateTime   time.Time `json:"createdDateTime"`
}

type azureAppRoleAssignment struct {
	Id         string `json:"id"`
	Name       string `json:"name"`
	Properties struct {
		PrincipalId      string `json:"principalId"`
		RoleDefinitionId string `json:"roleDefinitionId"`
		Scope            string `json:"string"`
	} `json:"properties"`
}

type azureRoleDefinition struct {
	Name       string `json:"roleName"`
	Properties struct {
		Scopes      []string `json:"assignableScopes"`
		Permissions []struct {
			Actions    []string `json:"actions"`
			NotActions []string `json:"notActions"`
		} `json:"permissions"`
	} `json:"properties"`
}

func (r *azureRoleDefinition) toEtlRole() *types.EtlRole {
	permissions := map[string][]string{}
	for _, scope := range r.Properties.Scopes {
		for _, perm := range r.Properties.Permissions {
			// TODO: Handle not actions?
			permissions[scope] = perm.Actions
		}
	}

	return &types.EtlRole{
		Name:        r.Name,
		Permissions: permissions,
	}
}

func (u *azureUser) toEtlUser() *types.EtlUser {
	email := u.Mail
	if email == "" && len(u.OtherMails) > 0 {
		email = u.OtherMails[0]
	}

	return &types.EtlUser{
		Username:    u.UserPrincipalName,
		Email:       email,
		FullName:    u.DisplayName,
		CreatedTime: &u.CreatedDateTime,
		Roles:       map[string]*types.EtlRole{},
	}
}

type EtlAzureConnectorUser struct {
	opts *EtlAzureOptions
}

func createAzureConnectorUser(opts *EtlAzureOptions) (*EtlAzureConnectorUser, error) {
	return &EtlAzureConnectorUser{
		opts: opts,
	}, nil
}

func (c *EtlAzureConnectorUser) getAllUsers() ([]azureUser, *connectors.EtlSourceInfo, error) {
	type ResponseBody struct {
		NextLink string      `json:"@odata.nextLink"`
		Value    []azureUser `json:"value"`
	}

	endpoint := fmt.Sprintf("%s/users?$select=displayName,userPrincipalName,mail,otherMails,createdDateTime,id", baseGraphUrl)
	responses := []ResponseBody{}
	source, err := azurePaginatedGet(c.opts.GraphClient, endpoint, &responses)
	if err != nil {
		return nil, nil, err
	}

	retUsers := []azureUser{}
	for _, resp := range responses {
		retUsers = append(retUsers, resp.Value...)
	}
	return retUsers, source, nil
}

func (c *EtlAzureConnectorUser) getUserAppRoleAssignments(u *azureUser) ([]azureAppRoleAssignment, *connectors.EtlSourceInfo, error) {
	type ResponseBody struct {
		NextLink string                   `json:"@odata.nextLink"`
		Value    []azureAppRoleAssignment `json:"value"`
	}

	endpoint := fmt.Sprintf("%s/subscriptions/%s/providers/Microsoft.Authorization/roleAssignments?api-version=2015-07-01&$filter=assignedTo('%s')", azureManagementUrl, c.opts.SubscriptionId, u.Id)
	responses := []ResponseBody{}
	source, err := azurePaginatedGet(c.opts.ManagementClient, endpoint, &responses)
	if err != nil {
		return nil, nil, err
	}

	ret := []azureAppRoleAssignment{}
	for _, resp := range responses {
		ret = append(ret, resp.Value...)
	}
	return ret, source, nil
}

func (c *EtlAzureConnectorUser) getRoleDefinition(definitionId string) (*azureRoleDefinition, *connectors.EtlSourceInfo, error) {
	endpoint := fmt.Sprintf("%s/%s?api-version=2015-07-01&", azureManagementUrl, definitionId)
	response := azureRoleDefinition{}
	source, err := azureGet(c.opts.ManagementClient, endpoint, &response)
	if err != nil {
		return nil, nil, err
	}

	return &response, source, nil
}

type azureGetUserAppRoleAssignmentJob struct {
	// Input
	User      *azureUser
	Connector *EtlAzureConnectorUser

	// Output
	Roles     *[]azureAppRoleAssignment
	OutSource chan *connectors.EtlSourceInfo
}

func (j *azureGetUserAppRoleAssignmentJob) Do() error {
	roles, source, err := j.Connector.getUserAppRoleAssignments(j.User)
	if err != nil {
		return err
	}

	j.OutSource <- source
	*j.Roles = roles
	return nil
}

type azureGetRoleDefinitionJob struct {
	// Input
	DefinitionId string
	Connector    *EtlAzureConnectorUser

	// Output
	Def       *azureRoleDefinition
	OutSource chan *connectors.EtlSourceInfo
}

func (j *azureGetRoleDefinitionJob) Do() error {
	def, source, err := j.Connector.getRoleDefinition(j.DefinitionId)
	if err != nil {
		return err
	}

	j.OutSource <- source
	*j.Def = *def
	return nil
}

func (c *EtlAzureConnectorUser) GetUserListing() ([]*types.EtlUser, *connectors.EtlSourceInfo, error) {

	finalSource := connectors.CreateSourceInfo()

	// Step 1: Get all users.
	azureUsers, userSource, err := c.getAllUsers()
	if err != nil {
		return nil, nil, err
	}
	finalSource.MergeWith(userSource)

	wg := sync.WaitGroup{}
	wg.Add(1)

	sourcesToMerge := make(chan *connectors.EtlSourceInfo)
	go func(source *connectors.EtlSourceInfo, input chan *connectors.EtlSourceInfo) {
		defer wg.Done()
		for s := range input {
			source.MergeWith(s)
		}
	}(finalSource, sourcesToMerge)

	// Step 2: Get roles assigned to each user
	perUserRoles := map[string]*[]azureAppRoleAssignment{}
	{

		pool := mt.NewTaskPool(10)

		for _, u := range azureUsers {
			roles := []azureAppRoleAssignment{}

			pool.AddJob(&azureGetUserAppRoleAssignmentJob{
				User:      &u,
				Connector: c,
				Roles:     &roles,
				OutSource: sourcesToMerge,
			})

			perUserRoles[u.UserPrincipalName] = &roles
		}

		err := pool.SyncExecute()
		if err != nil {
			return nil, nil, err
		}
	}

	// Step 3: For all unique roles, get the role definition.
	perRoleDefinitions := map[string]*azureRoleDefinition{}
	uniqueAppRoles := map[string]*azureAppRoleAssignment{}
	{
		for _, allRoles := range perUserRoles {
			for _, appRole := range *allRoles {
				_, ok := uniqueAppRoles[appRole.Id]
				if ok {
					continue
				}
				uniqueAppRoles[appRole.Id] = &appRole
			}
		}

		pool := mt.NewTaskPool(10)

		for _, role := range uniqueAppRoles {
			definition := azureRoleDefinition{}

			pool.AddJob(&azureGetRoleDefinitionJob{
				DefinitionId: role.Properties.RoleDefinitionId,
				Connector:    c,
				Def:          &definition,
				OutSource:    sourcesToMerge,
			})

			perRoleDefinitions[role.Id] = &definition
		}

		err := pool.SyncExecute()
		if err != nil {
			return nil, nil, err
		}
	}

	// Step 4: Convert all the data we collected to our standardized format.
	etlRoles := map[string]*types.EtlRole{}
	for _, role := range uniqueAppRoles {
		definition := perRoleDefinitions[role.Id]
		etlRoles[role.Id] = definition.toEtlRole()
	}

	retUsers := []*types.EtlUser{}
	for _, u := range azureUsers {
		etlUser := u.toEtlUser()
		for _, role := range *perUserRoles[u.UserPrincipalName] {
			etlRole := etlRoles[role.Id]
			etlUser.Roles[etlRole.Name] = etlRole
		}
		retUsers = append(retUsers, etlUser)
	}

	close(sourcesToMerge)
	wg.Wait()
	return retUsers, finalSource, nil
}
