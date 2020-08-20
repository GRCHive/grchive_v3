package aws

import (
	"gitlab.com/grchive/grchive-v3/shared/etl/connectors"
	"gitlab.com/grchive/grchive-v3/shared/utility/http"
)

type EtlAWSOptions struct {
	Client http_utility.HttpClient
}

type EtlAWSConnector struct {
	opts  *EtlAWSOptions
	users *EtlAWSConnectorUser
}

const iamBaseUrl string = "https://iam.amazonaws.com"

func (c *EtlAWSConnector) GetUserInterface() (connectors.EtlConnectorUserInterface, error) {
	return c.users, nil
}

func CreateAWSConnector(opts *EtlAWSOptions) (*EtlAWSConnector, error) {
	var err error
	ret := EtlAWSConnector{
		opts: opts,
	}
	ret.users, err = createAWSConnectorUser(opts)

	if err != nil {
		return nil, err
	}

	return &ret, nil
}
