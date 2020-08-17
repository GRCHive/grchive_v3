package gsuite

import (
	"gitlab.com/grchive/grchive-v3/shared/etl/connectors"
	"gitlab.com/grchive/grchive-v3/shared/utility/http"
)

type EtlGSuiteOptions struct {
	Client     http_utility.HttpClient
	CustomerId string
}

type EtlGSuiteConnector struct {
	opts  *EtlGSuiteOptions
	users *EtlGSuiteConnectorUser
}

const baseUrl string = "https://www.googleapis.com/admin"
const directoryUrl = "/directory/v1"

func (c *EtlGSuiteConnector) GetUserInterface() (connectors.EtlConnectorUserInterface, error) {
	return c.users, nil
}

func CreateGSuiteConnector(opts *EtlGSuiteOptions) (*EtlGSuiteConnector, error) {
	var err error
	ret := EtlGSuiteConnector{
		opts: opts,
	}
	ret.users, err = createGSuiteConnectorUser(opts)

	if err != nil {
		return nil, err
	}

	return &ret, nil
}
