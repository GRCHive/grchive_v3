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

type azureAppRoleAssignmentProperties struct {
	PrincipalId      string `json:"principalId"`
	RoleDefinitionId string `json:"roleDefinitionId"`
	Scope            string `json:"scope"`
}

type azureAppRoleAssignment struct {
	Id         string                           `json:"id"`
	Name       string                           `json:"name"`
	Properties azureAppRoleAssignmentProperties `json:"properties"`
}

type azureRolePermission struct {
	Actions    []string `json:"actions"`
	NotActions []string `json:"notActions"`
}

type azureRoleDefinitionProperties struct {
	Name        string                `json:"roleName"`
	Scopes      []string              `json:"assignableScopes"`
	Permissions []azureRolePermission `json:"permissions"`
}

type azureRoleDefinition struct {
	Properties azureRoleDefinitionProperties `json:"properties"`
}

type azureDirectoryRole struct {
	Id          string `json:"id"`
	DisplayName string `json:"displayName"`
}

type azureDirectoryObject struct {
	Id string `json:"id"`
}

func (r azureDirectoryRole) toEtlRole() *types.EtlRole {
	return &types.EtlRole{
		Name:        r.DisplayName,
		Permissions: map[string][]string{},
	}
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
		Name:        r.Properties.Name,
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

func CreateAzureConnectorUser(opts *EtlAzureOptions) (*EtlAzureConnectorUser, error) {
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

func (c *EtlAzureConnectorUser) getUserAppRoleAssignments(userId string) ([]azureAppRoleAssignment, *connectors.EtlSourceInfo, error) {
	type ResponseBody struct {
		NextLink string                   `json:"@odata.nextLink"`
		Value    []azureAppRoleAssignment `json:"value"`
	}

	endpoint := fmt.Sprintf("%s/subscriptions/%s/providers/Microsoft.Authorization/roleAssignments?api-version=2015-07-01&$filter=assignedTo('%s')", azureManagementUrl, c.opts.SubscriptionId, userId)
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
	UserId    string
	Connector *EtlAzureConnectorUser

	// Output
	Roles     *[]azureAppRoleAssignment
	OutSource chan *connectors.EtlSourceInfo
}

func (j *azureGetUserAppRoleAssignmentJob) Do() error {
	roles, source, err := j.Connector.getUserAppRoleAssignments(j.UserId)
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

func (c *EtlAzureConnectorUser) getPerUserAzureRoles(azureUsers []azureUser, outRoles map[string][]*types.EtlRole) (*connectors.EtlSourceInfo, error) {
	wg := sync.WaitGroup{}
	wg.Add(1)

	finalSource := connectors.CreateSourceInfo()
	sourcesToMerge := make(chan *connectors.EtlSourceInfo)
	go func(source *connectors.EtlSourceInfo, input chan *connectors.EtlSourceInfo) {
		defer wg.Done()
		for s := range input {
			source.MergeWith(s)
		}
	}(finalSource, sourcesToMerge)

	// Step 1: Get Azure roles assigned to each user
	perUserRoles := map[string]*[]azureAppRoleAssignment{}
	{

		pool := mt.NewTaskPool(10)

		for _, u := range azureUsers {
			roles := []azureAppRoleAssignment{}

			pool.AddJob(&azureGetUserAppRoleAssignmentJob{
				UserId:    u.Id,
				Connector: c,
				Roles:     &roles,
				OutSource: sourcesToMerge,
			})

			perUserRoles[u.UserPrincipalName] = &roles
		}

		err := pool.SyncExecute()
		if err != nil {
			return nil, err
		}
	}

	// Step 2: For all unique roles, get the role definition.
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
			return nil, err
		}
	}

	close(sourcesToMerge)
	wg.Wait()

	// Step 3: Convert all the data we collected to our standardized format.
	etlRoles := map[string]*types.EtlRole{}
	for _, role := range uniqueAppRoles {
		definition := perRoleDefinitions[role.Id]
		etlRoles[role.Id] = definition.toEtlRole()
	}

	// Step 4: Assign roles to user.
	for _, u := range azureUsers {
		userRoles, ok := outRoles[u.Id]
		if !ok {
			userRoles = []*types.EtlRole{}
		}

		for _, role := range *perUserRoles[u.UserPrincipalName] {
			etlRole := etlRoles[role.Id]
			userRoles = append(userRoles, etlRole)
		}
		outRoles[u.Id] = userRoles
	}

	return finalSource, nil
}

func (c *EtlAzureConnectorUser) listDirectoryRoles() ([]azureDirectoryRole, *connectors.EtlSourceInfo, error) {
	type ResponseBody struct {
		NextLink string               `json:"@odata.nextLink"`
		Value    []azureDirectoryRole `json:"value"`
	}

	endpoint := fmt.Sprintf("%s/directoryRoles", baseGraphUrl)
	responses := []ResponseBody{}
	source, err := azurePaginatedGet(c.opts.GraphClient, endpoint, &responses)
	if err != nil {
		return nil, nil, err
	}

	retRoles := []azureDirectoryRole{}
	for _, resp := range responses {
		retRoles = append(retRoles, resp.Value...)
	}
	return retRoles, source, nil
}

func (c *EtlAzureConnectorUser) listUsersInDirectoryRole(roleId string) ([]azureDirectoryObject, *connectors.EtlSourceInfo, error) {
	type ResponseBody struct {
		NextLink string                 `json:"@odata.nextLink"`
		Value    []azureDirectoryObject `json:"value"`
	}

	endpoint := fmt.Sprintf("%s/directoryRoles/%s/members", baseGraphUrl, roleId)
	responses := []ResponseBody{}
	source, err := azurePaginatedGet(c.opts.GraphClient, endpoint, &responses)
	if err != nil {
		return nil, nil, err
	}

	retObjects := []azureDirectoryObject{}
	for _, resp := range responses {
		retObjects = append(retObjects, resp.Value...)
	}
	return retObjects, source, nil
}

func (c *EtlAzureConnectorUser) getPerUserDirectoryRoles(outRoles map[string][]*types.EtlRole) (*connectors.EtlSourceInfo, error) {
	finalSource := connectors.CreateSourceInfo()
	// This is slightly inefficient since we're querying every role when there might not necessarily be a user in every role but
	// the Microsoft Graph API doesn't seem to expose any other way of getting this information.
	// Step 1: List all directory roles.
	roles, src, err := c.listDirectoryRoles()
	if err != nil {
		return nil, err
	}
	finalSource.MergeWith(src)

	// Step 2: For each directory role, list the users in it.
	for _, r := range roles {
		members, src, err := c.listUsersInDirectoryRole(r.Id)
		if err != nil {
			return nil, err
		}
		finalSource.MergeWith(src)

		etlRole := r.toEtlRole()
		for _, m := range members {
			roleList, ok := outRoles[m.Id]
			if !ok {
				roleList = []*types.EtlRole{}
			}

			roleList = append(roleList, etlRole)
			outRoles[m.Id] = roleList
		}
	}

	return finalSource, nil
}

func (c *EtlAzureConnectorUser) GetUserListing() ([]*types.EtlUser, *connectors.EtlSourceInfo, error) {
	finalSource := connectors.CreateSourceInfo()

	// Step 1: Get all users.
	azureUsers, userSource, err := c.getAllUsers()
	if err != nil {
		return nil, nil, err
	}
	finalSource.MergeWith(userSource)

	perUserRoles := map[string][]*types.EtlRole{}
	// Step 2: Get Azure Roles. This only takes place is an Azure Subscription Id is passed in.
	if c.opts.SubscriptionId != "" {
		src, err := c.getPerUserAzureRoles(azureUsers, perUserRoles)
		if err != nil {
			return nil, nil, err
		}
		finalSource.MergeWith(src)
	}

	// Step 3: Get Directory Roles.
	{
		src, err := c.getPerUserDirectoryRoles(perUserRoles)
		if err != nil {
			return nil, nil, err
		}
		finalSource.MergeWith(src)
	}

	retUsers := []*types.EtlUser{}
	for _, u := range azureUsers {
		etlUser := u.toEtlUser()
		for _, role := range perUserRoles[u.Id] {
			etlUser.Roles[role.Name] = role
		}
		retUsers = append(retUsers, etlUser)
	}

	return retUsers, finalSource, nil
}
