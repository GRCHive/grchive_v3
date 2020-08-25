package linode

import (
	"gitlab.com/grchive/grchive-v3/shared/etl/connectors"
	"gitlab.com/grchive/grchive-v3/shared/utility/http"
)

type EtlLinodeOptions struct {
	Client http_utility.HttpClient
}

type EtlLinodeConnector struct {
	opts  *EtlLinodeOptions
	users *EtlLinodeConnectorUser
}

type linodeMeta struct {
	Total int64 `json:"total"`
	Links struct {
		Next string `json:"next"`
		Prev string `json:"prev"`
	} `json:"links"`
}

const apiUrl string = "https://api.linode.com/v4"

func (c *EtlLinodeConnector) GetUserInterface() (connectors.EtlConnectorUserInterface, error) {
	return c.users, nil
}

func CreateLinodeConnector(opts *EtlLinodeOptions) (*EtlLinodeConnector, error) {
	var err error
	ret := EtlLinodeConnector{
		opts: opts,
	}
	ret.users, err = createLinodeConnectorUser(opts)

	if err != nil {
		return nil, err
	}

	return &ret, nil
}
