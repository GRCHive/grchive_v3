package bitbucket

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

type bitbucketUser struct {
	Permission string `json:"permission"`
	User       struct {
		Nickname    string `json:"nickname"`
		DisplayName string `json:"display_name"`
		AccountId   string `json:"account_id"`
	} `json:"user"`
}

func (g bitbucketUser) toEtlUser() *types.EtlUser {
	return &types.EtlUser{
		Username: g.User.AccountId,
		FullName: g.User.DisplayName,
		Roles: map[string]*types.EtlRole{
			g.Permission: &types.EtlRole{
				Name: g.Permission,
			},
		},
	}
}

type EtlBitbucketConnectorUser struct {
	opts *EtlBitbucketOptions
}

func createBitbucketConnectorUser(opts *EtlBitbucketOptions) (*EtlBitbucketConnectorUser, error) {
	return &EtlBitbucketConnectorUser{
		opts: opts,
	}, nil
}

func (c *EtlBitbucketConnectorUser) GetUserListing() ([]*types.EtlUser, *connectors.EtlSourceInfo, error) {
	ctx := context.Background()
	retUsers := []*types.EtlUser{}
	source := connectors.CreateSourceInfo()
	uniqueUsers := map[string]bool{}

	endpoint := fmt.Sprintf(
		"%s/workspaces/%s/permissions",
		baseUrl,
		c.opts.WorkspaceId,
	)

	for {
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
			return nil, nil, errors.New("Bitbucket User Listing API Error: " + string(bodyData))
		}

		responseBody := struct {
			Next   *string `json:"next"`
			Values []bitbucketUser
		}{}
		err = json.Unmarshal(bodyData, &responseBody)
		if err != nil {
			return nil, nil, err
		}

		if len(responseBody.Values) == 0 {
			break
		}

		added := 0
		for _, u := range responseBody.Values {
			if _, ok := uniqueUsers[u.User.AccountId]; ok {
				continue
			}

			retUsers = append(retUsers, u.toEtlUser())
			uniqueUsers[u.User.AccountId] = true
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

		if responseBody.Next == nil {
			break
		} else {
			endpoint = *responseBody.Next
		}
	}

	return retUsers, source, nil
}
