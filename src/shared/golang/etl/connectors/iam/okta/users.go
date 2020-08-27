package okta

import (
	"fmt"
	"gitlab.com/grchive/grchive-v3/shared/etl/connectors"
	"gitlab.com/grchive/grchive-v3/shared/etl/types"
	"gitlab.com/grchive/grchive-v3/shared/utility/mt"
	"sync"
	"time"
)

type oktaProfile struct {
	Login       string `json:"login"`
	DisplayName string `json:"displayName"`
	Email       string `json:"email"`
	FirstName   string `json:"firstName"`
	LastName    string `json:"lastName"`
}

func (p oktaProfile) FullName() string {
	if p.DisplayName != "" {
		return p.DisplayName
	}
	return fmt.Sprintf("%s %s", p.FirstName, p.LastName)
}

type oktaUser struct {
	Id          string      `json:"id"`
	Created     time.Time   `json:"created"`
	LastUpdated time.Time   `json:"lastUpdated"`
	Profile     oktaProfile `json:"profile"`
}

func (u oktaUser) toEtlUser() *types.EtlUser {
	return &types.EtlUser{
		Username:       u.Profile.Login,
		FullName:       u.Profile.FullName(),
		Email:          u.Profile.Email,
		CreatedTime:    &u.Created,
		LastChangeTime: &u.LastUpdated,
		Roles:          map[string]*types.EtlRole{},
	}
}

type oktaRole struct {
	Label string `json:"label"`
}

func (r oktaRole) toEtlRole() *types.EtlRole {
	return &types.EtlRole{
		Name: r.Label,
	}
}

type EtlOktaConnectorUser struct {
	opts *EtlOktaOptions
}

func createEtlUserFromOkta(user *oktaUser, roles []oktaRole) *types.EtlUser {
	retUser := user.toEtlUser()
	for _, r := range roles {
		etlRole := r.toEtlRole()
		retUser.Roles[etlRole.Name] = etlRole
	}
	return retUser
}

func createOktaConnectorUser(opts *EtlOktaOptions) (*EtlOktaConnectorUser, error) {
	return &EtlOktaConnectorUser{
		opts: opts,
	}, nil
}

func (c *EtlOktaConnectorUser) getOktaUsers() ([]oktaUser, *connectors.EtlSourceInfo, error) {
	endpoint := fmt.Sprintf("%s/users", c.opts.apiBaseUrl())
	pages := [][]oktaUser{}
	source, err := oktaPaginatedGet(c.opts.Client, endpoint, &pages)

	if err != nil {
		return nil, nil, err
	}

	users := []oktaUser{}
	for _, p := range pages {
		users = append(users, p...)
	}

	return users, source, nil
}

func (c *EtlOktaConnectorUser) getOktaRoles(userId string) ([]oktaRole, *connectors.EtlSourceInfo, error) {
	endpoint := fmt.Sprintf("%s/users/%s/roles", c.opts.apiBaseUrl(), userId)
	pages := [][]oktaRole{}
	source, err := oktaPaginatedGet(c.opts.Client, endpoint, &pages)

	if err != nil {
		return nil, nil, err
	}

	roles := []oktaRole{}
	for _, p := range pages {
		roles = append(roles, p...)
	}

	return roles, source, nil
}

type oktaGetOktaRolesJob struct {
	// Input
	UserId    string
	Connector *EtlOktaConnectorUser

	// Output
	Roles     *[]oktaRole
	OutSource chan *connectors.EtlSourceInfo
}

func (j *oktaGetOktaRolesJob) Do() error {
	roles, source, err := j.Connector.getOktaRoles(j.UserId)
	if err != nil {
		return err
	}

	j.OutSource <- source
	*j.Roles = roles
	return nil
}

func (c *EtlOktaConnectorUser) GetUserListing() ([]*types.EtlUser, *connectors.EtlSourceInfo, error) {
	finalSrc := connectors.CreateSourceInfo()

	users, userSrc, err := c.getOktaUsers()
	if err != nil {
		return nil, nil, err
	}
	finalSrc.MergeWith(userSrc)

	wg := sync.WaitGroup{}
	wg.Add(1)

	sourcesToMerge := make(chan *connectors.EtlSourceInfo)
	go func(source *connectors.EtlSourceInfo, input chan *connectors.EtlSourceInfo) {
		defer wg.Done()
		for s := range input {
			source.MergeWith(s)
		}
	}(finalSrc, sourcesToMerge)

	perUserRoles := map[string]*[]oktaRole{}

	{
		pool := mt.NewTaskPool(10)
		for _, u := range users {
			roles := []oktaRole{}
			pool.AddJob(&oktaGetOktaRolesJob{
				UserId:    u.Id,
				Connector: c,
				Roles:     &roles,
				OutSource: sourcesToMerge,
			})
			perUserRoles[u.Id] = &roles
		}

		err := pool.SyncExecute()
		if err != nil {
			return nil, nil, err
		}
	}

	retUsers := make([]*types.EtlUser, len(users))
	for idx, u := range users {
		retUsers[idx] = createEtlUserFromOkta(&u, *perUserRoles[u.Id])
	}

	close(sourcesToMerge)
	wg.Wait()
	return retUsers, finalSrc, nil
}
