package vultr

import (
	"fmt"
	"gitlab.com/grchive/grchive-v3/shared/etl/connectors"
	"gitlab.com/grchive/grchive-v3/shared/etl/types"
)

type EtlVultrConnectorUser struct {
	opts *EtlVultrOptions
}

type vultrUser struct {
	Name  string   `json:"name"`
	Email string   `json:"email"`
	Acls  []string `json:"acls"`
}

func (u vultrUser) toEtlUser() *types.EtlUser {
	return &types.EtlUser{
		Username: u.Email,
		FullName: u.Name,
		Email:    u.Email,
		Roles: map[string]*types.EtlRole{
			"Self": &types.EtlRole{
				Name: "Self",
				Permissions: map[string][]string{
					"Self": u.Acls,
				},
			},
		},
	}
}

func createVultrConnectorUser(opts *EtlVultrOptions) (*EtlVultrConnectorUser, error) {
	return &EtlVultrConnectorUser{
		opts: opts,
	}, nil
}

func (c *EtlVultrConnectorUser) getUsers() ([]vultrUser, *connectors.EtlSourceInfo, error) {
	// Users endpoint only returns sub-users so we need to also query the /account endpoint to
	// get the parent user (the one that generated the API key).
	endpoint := fmt.Sprintf("%s/users", apiUrl)

	type ResponseBody struct {
		Meta  vultrMeta   `json:"meta"`
		Users []vultrUser `json:"users"`
	}

	responses := []ResponseBody{}
	source, err := vultrPaginatedGet(c.opts.Client, endpoint, &responses)
	if err != nil {
		return nil, nil, err
	}

	retUsers := []vultrUser{}
	for _, resp := range responses {
		retUsers = append(retUsers, resp.Users...)
	}
	return retUsers, source, nil
}

func (c *EtlVultrConnectorUser) getAccountInfo() (*vultrUser, *connectors.EtlSourceInfo, error) {
	endpoint := fmt.Sprintf("%s/account", apiUrl)

	type ResponseBody struct {
		Account vultrUser `json:"account"`
	}

	body := ResponseBody{}
	source, err := vultrGet(c.opts.Client, endpoint, &body)
	if err != nil {
		return nil, nil, err
	}

	return &body.Account, source, nil
}

func (c *EtlVultrConnectorUser) GetUserListing() ([]*types.EtlUser, *connectors.EtlSourceInfo, error) {
	finalSource := connectors.CreateSourceInfo()
	users, userSource, err := c.getUsers()
	if err != nil {
		return nil, nil, err
	}
	finalSource.MergeWith(userSource)

	account, accountSource, err := c.getAccountInfo()
	if err != nil {
		return nil, nil, err
	}

	users = append(users, *account)
	finalSource.MergeWith(accountSource)

	retUsers := make([]*types.EtlUser, len(users))
	for idx, u := range users {
		retUsers[idx] = u.toEtlUser()
	}

	return retUsers, finalSource, nil
}
