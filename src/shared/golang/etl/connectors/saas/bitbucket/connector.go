package bitbucket

import (
	"gitlab.com/grchive/grchive-v3/shared/etl/connectors"
	"gitlab.com/grchive/grchive-v3/shared/utility/http"
)

type EtlBitbucketOptions struct {
	Client      http_utility.HttpClient
	WorkspaceId string
}

type EtlBitbucketConnector struct {
	opts  *EtlBitbucketOptions
	users *EtlBitbucketConnectorUser
}

const baseUrl string = "https://api.bitbucket.org/2.0"

func (c *EtlBitbucketConnector) GetUserInterface() (connectors.EtlConnectorUserInterface, error) {
	return c.users, nil
}

func CreateBitbucketConnector(opts *EtlBitbucketOptions) (*EtlBitbucketConnector, error) {
	var err error
	ret := EtlBitbucketConnector{
		opts: opts,
	}
	ret.users, err = createBitbucketConnectorUser(opts)

	if err != nil {
		return nil, err
	}

	return &ret, nil
}
