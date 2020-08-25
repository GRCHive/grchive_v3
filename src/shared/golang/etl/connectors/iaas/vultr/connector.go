package vultr

import (
	"gitlab.com/grchive/grchive-v3/shared/etl/connectors"
	"gitlab.com/grchive/grchive-v3/shared/utility/http"
)

type EtlVultrOptions struct {
	Client http_utility.HttpClient
}

type EtlVultrConnector struct {
	opts  *EtlVultrOptions
	users *EtlVultrConnectorUser
}

type vultrMeta struct {
	Total int64 `json:"total"`
	Links struct {
		Next string `json:"next"`
		Prev string `json:"prev"`
	} `json:"links"`
}

const apiUrl string = "https://api.vultr.com/v2"

func (c *EtlVultrConnector) GetUserInterface() (connectors.EtlConnectorUserInterface, error) {
	return c.users, nil
}

func CreateVultrConnector(opts *EtlVultrOptions) (*EtlVultrConnector, error) {
	var err error
	ret := EtlVultrConnector{
		opts: opts,
	}
	ret.users, err = createVultrConnectorUser(opts)

	if err != nil {
		return nil, err
	}

	return &ret, nil
}
