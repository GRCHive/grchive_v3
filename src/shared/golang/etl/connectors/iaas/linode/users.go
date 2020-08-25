package linode

import (
	"fmt"
	"gitlab.com/grchive/grchive-v3/shared/etl/connectors"
	"gitlab.com/grchive/grchive-v3/shared/etl/types"
	"gitlab.com/grchive/grchive-v3/shared/utility/mt"
	"sync"
)

type EtlLinodeConnectorUser struct {
	opts *EtlLinodeOptions
}

type linodeUser struct {
	Username   string `json:"username"`
	Email      string `json:"email"`
	Restricted bool   `json:"restricted"`
}

type linodeGlobalGrants struct {
	AddLinodes           bool   `json:"add_linodes"`
	AddLongview          bool   `json:"add_longview"`
	LongviewSubscription bool   `json:"longview_subscription"`
	AccountAccess        string `json:"account_access"`
	CancelAccount        bool   `json:"cancel_account"`
	AddDomains           bool   `json:"add_domains"`
	AddStackScripts      bool   `json:"add_stackscripts"`
	AddNodeBalancers     bool   `json:"add_nodebalancers"`
	AddImages            bool   `json:"add_images"`
	AddVolumes           bool   `json:"add_volumes"`
}

type linodeResourceGrants struct {
	Id          int64  `json:"id"`
	Permissions string `json:"permissions"`
	Label       string `json:"label"`
}

type linodeGrants struct {
	Global       linodeGlobalGrants     `json:"global"`
	Linode       []linodeResourceGrants `json:"linode"`
	Domain       []linodeResourceGrants `json:"domain"`
	NodeBalancer []linodeResourceGrants `json:"nodebalancer"`
	Image        []linodeResourceGrants `json:"image"`
	LongView     []linodeResourceGrants `json:"longview"`
	StackScript  []linodeResourceGrants `json:"stackscript"`
	Volume       []linodeResourceGrants `json:"volume"`
}

func (g linodeGlobalGrants) toEtlRole() *types.EtlRole {
	role := &types.EtlRole{
		Name:        "Global",
		Permissions: map[string][]string{},
	}

	if g.AccountAccess != "" {
		role.Permissions["AccountAccess"] = []string{g.AccountAccess}
	}

	perms := []string{}

	if g.AddLinodes {
		perms = append(perms, "AddLinodes")
	}

	if g.AddLongview {
		perms = append(perms, "AddLongview")
	}

	if g.LongviewSubscription {
		perms = append(perms, "LongviewSubscription")
	}

	if g.CancelAccount {
		perms = append(perms, "CancelAccount")
	}

	if g.AddDomains {
		perms = append(perms, "AddDomains")
	}

	if g.AddStackScripts {
		perms = append(perms, "AddStackScripts")
	}

	if g.AddNodeBalancers {
		perms = append(perms, "AddNodeBalancers")
	}

	if g.AddImages {
		perms = append(perms, "AddImages")
	}

	if g.AddVolumes {
		perms = append(perms, "AddVolumes")
	}

	role.Permissions["Grants"] = perms
	return role
}

func (g linodeGrants) resourceToEtlRole(name string, r []linodeResourceGrants) *types.EtlRole {
	perms := map[string][]string{}
	for _, grant := range r {
		if grant.Permissions == "" {
			continue
		}
		perms[grant.Label] = []string{grant.Permissions}
	}

	return &types.EtlRole{
		Name:        name,
		Permissions: perms,
	}
}

func (g linodeGrants) toEtlRoles() []*types.EtlRole {
	allRoles := []*types.EtlRole{
		g.Global.toEtlRole(),
		g.resourceToEtlRole("Linode", g.Linode),
		g.resourceToEtlRole("Domain", g.Domain),
		g.resourceToEtlRole("NodeBalancer", g.NodeBalancer),
		g.resourceToEtlRole("Image", g.Image),
		g.resourceToEtlRole("LongView", g.LongView),
		g.resourceToEtlRole("StackScript", g.StackScript),
		g.resourceToEtlRole("Volume", g.Volume),
	}
	return allRoles
}

func (u linodeUser) toEtlUser() *types.EtlUser {
	return &types.EtlUser{
		Username: u.Username,
		Email:    u.Email,
		Roles:    map[string]*types.EtlRole{},
	}
}

func createLinodeConnectorUser(opts *EtlLinodeOptions) (*EtlLinodeConnectorUser, error) {
	return &EtlLinodeConnectorUser{
		opts: opts,
	}, nil
}

func (c *EtlLinodeConnectorUser) getUsers() ([]linodeUser, *connectors.EtlSourceInfo, error) {
	// Users endpoint only returns sub-users so we need to also query the /account endpoint to
	// get the parent user (the one that generated the API key).
	endpoint := fmt.Sprintf("%s/account/users", apiUrl)

	type ResponseBody struct {
		Pages int64        `json:"pages"`
		Data  []linodeUser `json:"data"`
	}

	responses := []ResponseBody{}
	source, err := linodePaginatedGet(c.opts.Client, endpoint, &responses)
	if err != nil {
		return nil, nil, err
	}

	retUsers := []linodeUser{}
	for _, resp := range responses {
		retUsers = append(retUsers, resp.Data...)
	}
	return retUsers, source, nil
}

func (c *EtlLinodeConnectorUser) getUserGrants(username string) (*linodeGrants, *connectors.EtlSourceInfo, error) {
	// Users endpoint only returns sub-users so we need to also query the /account endpoint to
	// get the parent user (the one that generated the API key).
	endpoint := fmt.Sprintf("%s/account/users/%s/grants", apiUrl, username)

	grants := linodeGrants{}
	source, err := linodeGet(c.opts.Client, endpoint, &grants)
	if err != nil {
		return nil, nil, err
	}

	return &grants, source, nil
}

type lindodeGetUserGrantsJob struct {
	// Input
	User      *linodeUser
	Connector *EtlLinodeConnectorUser

	// Output
	Grants    *linodeGrants
	OutSource chan *connectors.EtlSourceInfo
}

func (j *lindodeGetUserGrantsJob) Do() error {
	grants, source, err := j.Connector.getUserGrants(j.User.Username)
	if err != nil {
		return err
	}

	j.OutSource <- source
	*j.Grants = *grants
	return nil
}

func (c *EtlLinodeConnectorUser) GetUserListing() ([]*types.EtlUser, *connectors.EtlSourceInfo, error) {
	finalSource := connectors.CreateSourceInfo()
	users, userSource, err := c.getUsers()
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

	perUserGrants := map[string]*linodeGrants{}
	{
		pool := mt.NewTaskPool(10)

		for _, u := range users {
			if !u.Restricted {
				continue
			}
			grants := linodeGrants{}

			pool.AddJob(&lindodeGetUserGrantsJob{
				User:      &u,
				Connector: c,
				Grants:    &grants,
				OutSource: sourcesToMerge,
			})

			perUserGrants[u.Username] = &grants
		}

		err := pool.SyncExecute()
		if err != nil {
			return nil, nil, err
		}
	}

	retUsers := make([]*types.EtlUser, len(users))
	for idx, u := range users {
		retUsers[idx] = u.toEtlUser()

		if u.Restricted {
			roles := perUserGrants[u.Username].toEtlRoles()
			for _, r := range roles {
				retUsers[idx].Roles[r.Name] = r
			}
		}
	}

	close(sourcesToMerge)
	wg.Wait()

	return retUsers, finalSource, nil
}
