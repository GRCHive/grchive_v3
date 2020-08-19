package cloudflare

import (
	"gitlab.com/grchive/grchive-v3/shared/etl/connectors"
	"gitlab.com/grchive/grchive-v3/shared/utility/http"
)

type EtlCloudflareOptions struct {
	Client    http_utility.HttpClient
	AccountId string
}

type EtlCloudflareConnector struct {
	opts  *EtlCloudflareOptions
	users *EtlCloudflareConnectorUser
}

const baseUrl string = "https://api.cloudflare.com/client/v4"

func (c *EtlCloudflareConnector) GetUserInterface() (connectors.EtlConnectorUserInterface, error) {
	return c.users, nil
}

func CreateCloudflareConnector(opts *EtlCloudflareOptions) (*EtlCloudflareConnector, error) {
	var err error
	ret := EtlCloudflareConnector{
		opts: opts,
	}
	ret.users, err = createCloudflareConnectorUser(opts)

	if err != nil {
		return nil, err
	}

	return &ret, nil
}
