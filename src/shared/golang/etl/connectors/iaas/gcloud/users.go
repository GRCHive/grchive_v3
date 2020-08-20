package gcloud

import (
	"encoding/json"
	"errors"
	"fmt"
	"gitlab.com/grchive/grchive-v3/shared/etl/connectors"
	"gitlab.com/grchive/grchive-v3/shared/etl/types"
	"gitlab.com/grchive/grchive-v3/shared/utility/mt"
	"golang.org/x/net/context"
	"io/ioutil"
	"net/http"
	"strings"
)

type gcloudRole struct {
	Name                string   `json:"name"`
	Title               string   `json:"title"`
	IncludedPermissions []string `json:"includedPermissions"`
}

type gcloudIamPolicy struct {
	Bindings []struct {
		Role    string   `json:"role"`
		Members []string `json:"members"`
	} `json:"bindings"`
}

type EtlGCloudConnectorUser struct {
	opts *EtlGCloudOptions
}

type populateGCloudRolePermissionsJob struct {
	role      *types.EtlRole
	outCmds   chan *connectors.EtlSourceInfo
	connector *EtlGCloudConnectorUser
}

func (j *populateGCloudRolePermissionsJob) Do() error {
	role, cmd, err := j.connector.getCloudRole(j.role.Name)
	if err != nil {
		return err
	}

	j.role.Permissions["Self"] = role.IncludedPermissions
	j.role.Name = role.Title
	j.outCmds <- cmd
	return nil
}

func createGCloudConnectorUser(opts *EtlGCloudOptions) (*EtlGCloudConnectorUser, error) {
	return &EtlGCloudConnectorUser{
		opts: opts,
	}, nil
}

func (c *EtlGCloudConnectorUser) getCloudRole(role string) (*gcloudRole, *connectors.EtlSourceInfo, error) {
	ctx := context.Background()
	endpoint := fmt.Sprintf(
		"%s/v1/%s",
		iamBaseUrl,
		role,
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
		return nil, nil, errors.New("GCloud Get Role API Error: " + string(bodyData))
	}

	body := gcloudRole{}
	err = json.Unmarshal(bodyData, &body)
	if err != nil {
		return nil, nil, err
	}

	source := connectors.CreateSourceInfo()
	cmd := connectors.EtlCommandInfo{
		Command: endpoint,
		RawData: string(bodyData),
	}
	source.AddCommand(&cmd)
	return &body, source, nil
}

func (c *EtlGCloudConnectorUser) GetUserListing() ([]*types.EtlUser, *connectors.EtlSourceInfo, error) {
	ctx := context.Background()
	retUsers := []*types.EtlUser{}
	source := connectors.CreateSourceInfo()

	endpoint := fmt.Sprintf(
		"%s/v1/projects/%s:getIamPolicy",
		resourceManagerBaseUrl,
		c.opts.ProjectId,
	)

	req, err := http.NewRequestWithContext(
		ctx,
		"POST",
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
		return nil, nil, errors.New("GCloud User Listing API Error: " + string(bodyData))
	}

	body := gcloudIamPolicy{}
	err = json.Unmarshal(bodyData, &body)
	if err != nil {
		return nil, nil, err
	}

	// Need to do a two things here.
	// 	1) Parse out the given bindings so that we know all the users and service accounts that have
	// 	   roles in this project.
	// 	2) For each unique role, determine which permissions that role has.
	taskPool := mt.NewTaskPool(10)
	allUsers := map[string]*types.EtlUser{}

	roleCmds := make(chan *connectors.EtlSourceInfo, len(body.Bindings))
	for _, binding := range body.Bindings {
		newRole := &types.EtlRole{
			Name: binding.Role,
			Permissions: map[string][]string{
				"Self": []string{},
			},
		}

		taskPool.AddJob(&populateGCloudRolePermissionsJob{
			role:      newRole,
			outCmds:   roleCmds,
			connector: c,
		})

		for _, u := range binding.Members {
			etlUser, ok := allUsers[u]
			if !ok {
				split := strings.Split(u, ":")

				etlUser = &types.EtlUser{
					Username: u,
					Email:    split[1],
					Roles:    map[string]*types.EtlRole{},
				}

				allUsers[u] = etlUser
			}

			etlUser.Roles[newRole.Name] = newRole
		}
	}

	err = taskPool.SyncExecute()
	close(roleCmds)
	if err != nil {
		return nil, nil, err
	}

	for _, v := range allUsers {
		retUsers = append(retUsers, v)
	}

	cmd := connectors.EtlCommandInfo{
		Command: endpoint,
		RawData: string(bodyData),
	}
	source.AddCommand(&cmd)

	for s := range roleCmds {
		source.MergeWith(s)
	}
	return retUsers, source, nil
}
