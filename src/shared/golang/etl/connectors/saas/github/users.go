package github

import (
	"encoding/json"
	"errors"
	"fmt"
	"gitlab.com/grchive/grchive-v3/shared/etl/connectors"
	"gitlab.com/grchive/grchive-v3/shared/etl/types"
	"golang.org/x/net/context"
	"io/ioutil"
	"net/http"
)

type githubUser struct {
	Login string `json:"login"`
}

type EtlGithubConnectorUser struct {
	opts *EtlGithubOptions
}

func (g githubUser) toEtlUser(role string) *types.EtlUser {
	roles := map[string]*types.EtlRole{}
	roles[role] = &types.EtlRole{
		Name:        role,
		Permissions: map[string][]string{},
	}

	return &types.EtlUser{
		Username: g.Login,
		Roles:    roles,
	}
}

func createGithubConnectorUser(opts *EtlGithubOptions) (*EtlGithubConnectorUser, error) {
	return &EtlGithubConnectorUser{
		opts: opts,
	}, nil
}

func (c *EtlGithubConnectorUser) getUserListingHelper(role string) ([]*types.EtlUser, *connectors.EtlSourceInfo, error) {
	ctx := context.Background()
	retUsers := []*types.EtlUser{}
	source := connectors.CreateSourceInfo()

	page := 1
	const perPage int = 100

	uniqueUsers := map[string]bool{}

	for {
		endpoint := fmt.Sprintf(
			"%s/orgs/%s/members?role=%s&page=%d&per_page=%d",
			baseUrl,
			c.opts.OrgId,
			role,
			page,
			perPage,
		)

		req, err := http.NewRequestWithContext(
			ctx,
			"GET",
			endpoint,
			nil,
		)
		if err != nil {
			return nil, nil, err
		}

		resp, err := c.opts.Client.Do(req)
		if err != nil {
			return nil, nil, err
		}
		defer resp.Body.Close()

		bodyData, err := ioutil.ReadAll(resp.Body)
		if err != nil {
			return nil, nil, err
		}

		if resp.StatusCode != http.StatusOK {
			return nil, nil, errors.New("Github User Listing API Error:" + string(bodyData))
		}

		body := []githubUser{}
		err = json.Unmarshal(bodyData, &body)
		if err != nil {
			return nil, nil, err
		}

		if len(body) == 0 {
			break
		}

		added := 0
		for _, u := range body {
			newU := u.toEtlUser(role)
			if _, ok := uniqueUsers[newU.Username]; ok {
				continue
			}
			uniqueUsers[newU.Username] = true
			retUsers = append(retUsers, newU)
			added = added + 1
		}

		if added == 0 {
			break
		}

		cmd := connectors.EtlCommandInfo{
			Command: endpoint,
			RawData: string(bodyData),
		}
		source.AddCommand(&cmd)

		page = page + 1
	}
	return retUsers, source, nil
}

func (c *EtlGithubConnectorUser) GetUserListing() ([]*types.EtlUser, *connectors.EtlSourceInfo, error) {
	// Get users in two passes - one to get admins and one to get members.
	admins, adminSrc, err := c.getUserListingHelper("admin")
	if err != nil {
		return nil, nil, err
	}

	members, memberSrc, err := c.getUserListingHelper("member")
	if err != nil {
		return nil, nil, err
	}

	adminSrc.MergeWith(memberSrc)
	return append(admins, members...), adminSrc, nil
}
