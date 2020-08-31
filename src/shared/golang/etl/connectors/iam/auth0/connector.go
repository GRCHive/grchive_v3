package auth0

import (
	"fmt"
	"gitlab.com/grchive/grchive-v3/shared/etl/connectors"
	"gitlab.com/grchive/grchive-v3/shared/utility/http"
)

type EtlAuth0Options struct {
	Client http_utility.HttpClient
	Domain string
}

func (o EtlAuth0Options) apiBaseUrl() string {
	return fmt.Sprintf("https://%s/api/v2", o.Domain)
}

type EtlAuth0Connector struct {
	opts  *EtlAuth0Options
	users *EtlAuth0ConnectorUser
}

func (c *EtlAuth0Connector) GetUserInterface() (connectors.EtlConnectorUserInterface, error) {
	return c.users, nil
}

func CreateAuth0Connector(opts *EtlAuth0Options) (*EtlAuth0Connector, error) {
	var err error
	ret := EtlAuth0Connector{
		opts: opts,
	}
	ret.users, err = createAuth0ConnectorUser(opts)

	if err != nil {
		return nil, err
	}

	return &ret, nil
}
