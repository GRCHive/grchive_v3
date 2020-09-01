package v5

import (
	"gitlab.com/grchive/grchive-v3/shared/etl/connectors"
	"gitlab.com/grchive/grchive-v3/shared/etl/connectors/databases"
)

type InterfaceFactory struct {
}

func (f *InterfaceFactory) CreateUserInterface(db *databases.DB) (connectors.EtlConnectorUserInterface, error) {
	return &EtlMysqlV5ConnectorUser{db: db}, nil
}
