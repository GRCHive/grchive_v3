package heroku

import (
	"gitlab.com/grchive/grchive-v3/shared/etl/connectors"
	"gitlab.com/grchive/grchive-v3/shared/utility/http"
)

type EtlHerokuOptions struct {
	Client   http_utility.HttpClient
	TeamName string
}

type EtlHerokuConnector struct {
	opts  *EtlHerokuOptions
	users *EtlHerokuConnectorUser
}

const apiUrl string = "https://api.heroku.com"

func (c *EtlHerokuConnector) GetUserInterface() (connectors.EtlConnectorUserInterface, error) {
	return c.users, nil
}

func CreateHerokuConnector(opts *EtlHerokuOptions) (*EtlHerokuConnector, error) {
	var err error
	ret := EtlHerokuConnector{
		opts: opts,
	}
	ret.users, err = createHerokuConnectorUser(opts)

	if err != nil {
		return nil, err
	}

	return &ret, nil
}
