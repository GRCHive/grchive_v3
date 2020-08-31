package auth0

import (
	"fmt"
	"gitlab.com/grchive/grchive-v3/shared/etl/connectors"
	"gitlab.com/grchive/grchive-v3/shared/etl/types"
	"time"
)

type auth0User struct {
	Name      string    `json:"name"`
	Email     string    `json:"email"`
	CreatedAt time.Time `json:"created_at"`
}

func (u auth0User) toEtlUser() *types.EtlUser {
	return &types.EtlUser{
		Username:    u.Email,
		Email:       u.Email,
		FullName:    u.Name,
		CreatedTime: &u.CreatedAt,
		Roles:       map[string]*types.EtlRole{},
	}
}

type EtlAuth0ConnectorUser struct {
	opts *EtlAuth0Options
}

func createAuth0ConnectorUser(opts *EtlAuth0Options) (*EtlAuth0ConnectorUser, error) {
	return &EtlAuth0ConnectorUser{
		opts: opts,
	}, nil
}

func (c *EtlAuth0ConnectorUser) GetUserListing() ([]*types.EtlUser, *connectors.EtlSourceInfo, error) {
	endpoint := fmt.Sprintf("%s/users", c.opts.apiBaseUrl())

	pages := [][]auth0User{}
	source, err := auth0PaginatedGet(c.opts.Client, endpoint, &pages)
	if err != nil {
		return nil, nil, err
	}

	retUsers := []*types.EtlUser{}
	for _, page := range pages {
		for _, u := range page {
			retUsers = append(retUsers, u.toEtlUser())
		}
	}
	return retUsers, source, nil
}
