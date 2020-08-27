package okta

import (
	"fmt"
	"gitlab.com/grchive/grchive-v3/shared/etl/connectors"
	"gitlab.com/grchive/grchive-v3/shared/utility/http"
)

type EtlOktaOptions struct {
	Client http_utility.HttpClient
	Domain string
}

func (o EtlOktaOptions) apiBaseUrl() string {
	return fmt.Sprintf("https://%s/api/v1", o.Domain)
}

type EtlOktaConnector struct {
	opts  *EtlOktaOptions
	users *EtlOktaConnectorUser
}

func (c *EtlOktaConnector) GetUserInterface() (connectors.EtlConnectorUserInterface, error) {
	return c.users, nil
}

func CreateOktaConnector(opts *EtlOktaOptions) (*EtlOktaConnector, error) {
	var err error
	ret := EtlOktaConnector{
		opts: opts,
	}
	ret.users, err = createOktaConnectorUser(opts)

	if err != nil {
		return nil, err
	}

	return &ret, nil
}
