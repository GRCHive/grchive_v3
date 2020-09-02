package office365

import (
	"gitlab.com/grchive/grchive-v3/shared/etl/connectors"
	"gitlab.com/grchive/grchive-v3/shared/etl/connectors/iaas/azure"
	"gitlab.com/grchive/grchive-v3/shared/utility/http"
)

type EtlOffice365Options struct {
	Client http_utility.HttpClient
}

type EtlOffice365Connector struct {
	opts  *EtlOffice365Options
	users connectors.EtlConnectorUserInterface
}

func (c *EtlOffice365Connector) GetUserInterface() (connectors.EtlConnectorUserInterface, error) {
	return c.users, nil
}

func CreateOffice365Connector(opts *EtlOffice365Options) (*EtlOffice365Connector, error) {
	var err error
	ret := EtlOffice365Connector{
		opts: opts,
	}
	ret.users, err = azure.CreateAzureConnectorUser(&azure.EtlAzureOptions{
		GraphClient: opts.Client,
	})

	if err != nil {
		return nil, err
	}

	return &ret, nil
}
