package heroku

import (
	"fmt"
	"gitlab.com/grchive/grchive-v3/shared/etl/connectors"
	"gitlab.com/grchive/grchive-v3/shared/etl/types"
	"time"
)

type herokuUser struct {
	Id    string `json:"id"`
	Email string `json:"email"`
	Name  string `json:"name"`
}

type herokuTeamMember struct {
	Role      string     `json:"role"`
	CreatedAt time.Time  `json:"created_at"`
	User      herokuUser `json:"user"`
}

func (m herokuTeamMember) toEtlUser() *types.EtlUser {
	return &types.EtlUser{
		Username:    m.User.Email,
		Email:       m.User.Email,
		FullName:    m.User.Name,
		CreatedTime: &m.CreatedAt,
		Roles: map[string]*types.EtlRole{
			m.Role: &types.EtlRole{
				Name:        m.Role,
				Permissions: map[string][]string{},
			},
		},
	}
}

type EtlHerokuConnectorUser struct {
	opts *EtlHerokuOptions
}

func createHerokuConnectorUser(opts *EtlHerokuOptions) (*EtlHerokuConnectorUser, error) {
	return &EtlHerokuConnectorUser{
		opts: opts,
	}, nil
}

func (c *EtlHerokuConnectorUser) GetUserListing() ([]*types.EtlUser, *connectors.EtlSourceInfo, error) {
	endpoint := fmt.Sprintf("%s/teams/%s/members", apiUrl, c.opts.TeamName)

	pages := [][]herokuTeamMember{}
	source, err := herokuPaginatedGet(c.opts.Client, endpoint, &pages)
	if err != nil {
		return nil, nil, err
	}

	retUsers := []*types.EtlUser{}
	for _, members := range pages {
		for _, u := range members {
			retUsers = append(retUsers, u.toEtlUser())
		}
	}

	return retUsers, source, nil
}
