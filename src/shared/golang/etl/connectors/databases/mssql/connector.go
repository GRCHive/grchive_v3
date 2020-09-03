package mssql

import (
	"gitlab.com/grchive/grchive-v3/shared/etl/connectors"
	"gitlab.com/grchive/grchive-v3/shared/etl/connectors/databases"
)

type EtlMssqlConnector struct {
	db    databases.SqlxLike
	users *EtlMssqlConnectorUser
}

func (c *EtlMssqlConnector) GetUserInterface() (connectors.EtlConnectorUserInterface, error) {
	return c.users, nil
}

func CreateMssqlConnector(db databases.SqlxLike) (*EtlMssqlConnector, error) {
	var err error
	ret := EtlMssqlConnector{
		db: db,
	}
	ret.users, err = createMssqlConnectorUser(&databases.DB{
		SqlxLike: db,
	})
	if err != nil {
		return nil, err
	}

	return &ret, nil
}
