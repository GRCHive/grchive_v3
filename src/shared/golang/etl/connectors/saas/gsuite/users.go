package gsuite

import (
	"encoding/json"
	"errors"
	"fmt"
	"gitlab.com/grchive/grchive-v3/shared/etl/connectors"
	"gitlab.com/grchive/grchive-v3/shared/etl/types"
	"golang.org/x/net/context"
	"io/ioutil"
	"net/http"
	"time"
)

type gsuiteUser struct {
	Kind         string `json:"kind"`
	Id           string `json:"id"`
	PrimaryEmail string `json:"primaryEmail"`
	Name         struct {
		GivenName  string `json:"givenName"`
		FamilyName string `json:"familyName"`
		FullName   string `json:"fullName"`
	} `json:"name"`
	CreationTime     time.Time `json:"creationTime"`
	IsAdmin          bool      `json:"isAdmin"`
	IsDelegatedAdmin bool      `json:"isDelegatedAdmin"`
}

const adminRole = "admin"
const delegatedAdminRole = "delegatedAdmin"

func (g gsuiteUser) toEtlUser() *types.EtlUser {
	roles := map[string]*types.EtlRole{}

	if g.IsAdmin {
		roles[adminRole] = &types.EtlRole{
			Name:        adminRole,
			Permissions: map[string][]string{},
		}
	}

	if g.IsDelegatedAdmin {
		roles[delegatedAdminRole] = &types.EtlRole{
			Name:        delegatedAdminRole,
			Permissions: map[string][]string{},
		}
	}

	return &types.EtlUser{
		Username:    g.PrimaryEmail,
		FullName:    g.Name.FullName,
		Email:       g.PrimaryEmail,
		CreatedTime: &g.CreationTime,
		Roles:       roles,
	}
}

type EtlGSuiteConnectorUser struct {
	opts *EtlGSuiteOptions
}

func createGSuiteConnectorUser(opts *EtlGSuiteOptions) (*EtlGSuiteConnectorUser, error) {
	return &EtlGSuiteConnectorUser{
		opts: opts,
	}, nil
}

func (c *EtlGSuiteConnectorUser) GetUserListing() ([]*types.EtlUser, *connectors.EtlSourceInfo, error) {
	ctx := context.Background()
	retUsers := []*types.EtlUser{}

	emptyNextPageToken := ""
	var nextPageToken *string
	nextPageToken = &emptyNextPageToken

	source := connectors.CreateSourceInfo()

	for nextPageToken != nil {
		endpoint := fmt.Sprintf(
			"%s%s/users?customer=%s",
			baseUrl,
			directoryUrl,
			c.opts.CustomerId,
		)

		if *nextPageToken != "" {
			endpoint = endpoint + "&pageToken=" + *nextPageToken
		}

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

		if resp.StatusCode != http.StatusOK {
			return nil, nil, errors.New("GSuite User Listing API Error.")
		}

		type responseBody struct {
			Kind          string       `json:"kind"`
			Users         []gsuiteUser `json:"users"`
			NextPageToken *string      `json:"nextPageToken"`
		}

		bodyData, err := ioutil.ReadAll(resp.Body)
		if err != nil {
			return nil, nil, err
		}

		body := responseBody{}
		err = json.Unmarshal(bodyData, &body)
		if err != nil {
			return nil, nil, err
		}

		for _, u := range body.Users {
			retUsers = append(retUsers, u.toEtlUser())
		}

		nextPageToken = body.NextPageToken
		cmd := connectors.EtlCommandInfo{
			Command: endpoint,
			RawData: string(bodyData),
		}
		source.AddCommand(&cmd)
	}

	return retUsers, source, nil
}
