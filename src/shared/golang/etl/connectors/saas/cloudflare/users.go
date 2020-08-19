package cloudflare

import (
	"encoding/json"
	"errors"
	"fmt"
	"gitlab.com/grchive/grchive-v3/shared/etl/connectors"
	"gitlab.com/grchive/grchive-v3/shared/etl/types"
	"golang.org/x/net/context"
	"io/ioutil"
	"net/http"
	"strings"
)

const readPermission = "Read"
const editPermission = "Edit"
const writePermission = "Write"

type cloudflareUser struct {
	User struct {
		FirstName string `json:"first_name"`
		LastName  string `json:"last_name"`
		Email     string `json:"email"`
	} `json:"user"`
	Roles []struct {
		Name        string `json:"name"`
		Permissions map[string]struct {
			Read  bool `json:"read"`
			Edit  bool `json:"edit"`
			Write bool `json:"write"`
		} `json:"permissions"`
	} `json:"roles"`
}

func (g cloudflareUser) toEtlUser() *types.EtlUser {
	roles := map[string]*types.EtlRole{}
	for _, r := range g.Roles {
		etlRole := &types.EtlRole{
			Name:        r.Name,
			Permissions: map[string][]string{},
		}

		for nm, perm := range r.Permissions {
			newPermissions := []string{}

			if perm.Read {
				newPermissions = append(newPermissions, readPermission)
			}

			if perm.Edit {
				newPermissions = append(newPermissions, editPermission)
			}

			if perm.Write {
				newPermissions = append(newPermissions, writePermission)
			}

			etlRole.Permissions[nm] = newPermissions
		}

		roles[r.Name] = etlRole
	}

	return &types.EtlUser{
		Username: g.User.Email,
		Email:    g.User.Email,
		FullName: strings.TrimSpace(g.User.FirstName + " " + g.User.LastName),
		Roles:    roles,
	}
}

type EtlCloudflareConnectorUser struct {
	opts *EtlCloudflareOptions
}

func createCloudflareConnectorUser(opts *EtlCloudflareOptions) (*EtlCloudflareConnectorUser, error) {
	return &EtlCloudflareConnectorUser{
		opts: opts,
	}, nil
}

func (c *EtlCloudflareConnectorUser) GetUserListing() ([]*types.EtlUser, *connectors.EtlSourceInfo, error) {
	ctx := context.Background()
	retUsers := []*types.EtlUser{}
	source := connectors.CreateSourceInfo()
	uniqueUsers := map[string]bool{}

	page := 1

	for {
		endpoint := fmt.Sprintf(
			"%s/accounts/%s/members?page=%d&per_page=50&direction=desc",
			baseUrl,
			c.opts.AccountId,
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
			return nil, nil, errors.New("Cloudflare User Listing API Error: " + string(bodyData))
		}

		responseBody := struct {
			Success    bool             `json:"success"`
			Errors     []string         `json:"errors"`
			Result     []cloudflareUser `json:"result"`
			ResultInfo struct {
				Page       int `json:"page"`
				TotalPages int `json:"total_pages"`
			} `json:"result_info"`
		}{}
		err = json.Unmarshal(bodyData, &responseBody)
		if err != nil {
			return nil, nil, err
		}

		if !responseBody.Success {
			return nil, nil, errors.New(strings.Join(responseBody.Errors, "\n"))
		}

		if len(responseBody.Result) == 0 {
			break
		}

		added := 0
		for _, u := range responseBody.Result {
			if _, ok := uniqueUsers[u.User.Email]; ok {
				continue
			}

			retUsers = append(retUsers, u.toEtlUser())
			uniqueUsers[u.User.Email] = true
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

		page += 1

		if responseBody.ResultInfo.Page == responseBody.ResultInfo.TotalPages {
			break
		}
	}

	return retUsers, source, nil
}
