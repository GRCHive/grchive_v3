package gcloud

import (
	"gitlab.com/grchive/grchive-v3/shared/etl/connectors"
	"gitlab.com/grchive/grchive-v3/shared/utility/http"
)

type EtlGCloudOptions struct {
	Client    http_utility.HttpClient
	ProjectId string
}

type EtlGCloudConnector struct {
	opts  *EtlGCloudOptions
	users *EtlGCloudConnectorUser
}

const resourceManagerBaseUrl string = "https://cloudresourcemanager.googleapis.com"
const iamBaseUrl string = "https://iam.googleapis.com"

func (c *EtlGCloudConnector) GetUserInterface() (connectors.EtlConnectorUserInterface, error) {
	return c.users, nil
}

func CreateGCloudConnector(opts *EtlGCloudOptions) (*EtlGCloudConnector, error) {
	var err error
	ret := EtlGCloudConnector{
		opts: opts,
	}
	ret.users, err = createGCloudConnectorUser(opts)

	if err != nil {
		return nil, err
	}

	return &ret, nil
}
