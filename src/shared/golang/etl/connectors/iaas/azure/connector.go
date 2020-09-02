package azure

import (
	"gitlab.com/grchive/grchive-v3/shared/etl/connectors"
	"gitlab.com/grchive/grchive-v3/shared/utility/http"
)

type EtlAzureOptions struct {
	ManagementClient http_utility.HttpClient
	GraphClient      http_utility.HttpClient
	SubscriptionId   string
}

type EtlAzureConnector struct {
	opts  *EtlAzureOptions
	users *EtlAzureConnectorUser
}

const baseGraphUrl = "https://graph.microsoft.com/v1.0"
const azureManagementUrl = "https://management.azure.com"

func (c *EtlAzureConnector) GetUserInterface() (connectors.EtlConnectorUserInterface, error) {
	return c.users, nil
}

func CreateAzureConnector(opts *EtlAzureOptions) (*EtlAzureConnector, error) {
	var err error
	ret := EtlAzureConnector{
		opts: opts,
	}
	ret.users, err = CreateAzureConnectorUser(opts)

	if err != nil {
		return nil, err
	}

	return &ret, nil
}
