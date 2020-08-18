package gitlab

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

type gitlabUser struct {
	Id          int64       `json:"id"`
	Username    string      `json:"username"`
	Name        string      `json:"name"`
	AccessLevel AccessLevel `json:"access_level"`
}

func (g gitlabUser) toEtlUser() *types.EtlUser {
	return &types.EtlUser{
		Username: g.Username,
		FullName: g.Name,
		Roles: map[string]*types.EtlRole{
			g.AccessLevel.ToString(): &types.EtlRole{
				Name: g.AccessLevel.ToString(),
			},
		},
	}

}

type EtlGitlabConnectorUser struct {
	opts *EtlGitlabOptions
}

func createGitlabConnectorUser(opts *EtlGitlabOptions) (*EtlGitlabConnectorUser, error) {
	return &EtlGitlabConnectorUser{
		opts: opts,
	}, nil
}

func (c *EtlGitlabConnectorUser) GetUserListing() ([]*types.EtlUser, *connectors.EtlSourceInfo, error) {
	ctx := context.Background()
	retUsers := []*types.EtlUser{}
	source := connectors.CreateSourceInfo()
	uniqueUsers := map[int64]bool{}

	page := 1
	for {
		endpoint := fmt.Sprintf(
			"%s/groups/%s/members/all?page=%d&per_page=100",
			baseUrl,
			c.opts.GroupId,
			page,
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
			return nil, nil, errors.New("Gitlab User Listing API Error: " + string(bodyData))
		}

		gitlabUsers := []gitlabUser{}
		err = json.Unmarshal(bodyData, &gitlabUsers)
		if err != nil {
			return nil, nil, err
		}

		if len(gitlabUsers) == 0 {
			break
		}

		added := 0
		for _, u := range gitlabUsers {
			if _, ok := uniqueUsers[u.Id]; ok {
				continue
			}

			retUsers = append(retUsers, u.toEtlUser())
			uniqueUsers[u.Id] = true
			added = added + 1
		}

		if added == 0 {
			break
		}

		page = page + 1
		cmd := connectors.EtlCommandInfo{
			Command: endpoint,
			RawData: string(bodyData),
		}
		source.AddCommand(&cmd)
	}

	return retUsers, source, nil
}
