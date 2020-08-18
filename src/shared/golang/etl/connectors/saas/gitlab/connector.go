package gitlab

import (
	"gitlab.com/grchive/grchive-v3/shared/etl/connectors"
	"gitlab.com/grchive/grchive-v3/shared/utility/http"
)

type EtlGitlabOptions struct {
	Client  http_utility.HttpClient
	GroupId string
}

type EtlGitlabConnector struct {
	opts  *EtlGitlabOptions
	users *EtlGitlabConnectorUser
}

const baseUrl string = "https://gitlab.com/api/v4"

func (c *EtlGitlabConnector) GetUserInterface() (connectors.EtlConnectorUserInterface, error) {
	return c.users, nil
}

func CreateGitlabConnector(opts *EtlGitlabOptions) (*EtlGitlabConnector, error) {
	var err error
	ret := EtlGitlabConnector{
		opts: opts,
	}
	ret.users, err = createGitlabConnectorUser(opts)

	if err != nil {
		return nil, err
	}

	return &ret, nil
}
