package psql

import (
	"gitlab.com/grchive/grchive-v3/shared/etl/connectors"
	"gitlab.com/grchive/grchive-v3/shared/etl/connectors/databases"
)

type EtlPsqlConnector struct {
	db    databases.SqlxLike
	users *EtlPsqlConnectorUser
}

func (c *EtlPsqlConnector) GetUserInterface() (connectors.EtlConnectorUserInterface, error) {
	return c.users, nil
}

func CreatePsqlConnector(db databases.SqlxLike) (*EtlPsqlConnector, error) {
	var err error
	ret := EtlPsqlConnector{
		db: db,
	}
	ret.users, err = createPsqlConnectorUser(&databases.DB{
		SqlxLike: db,
	})
	if err != nil {
		return nil, err
	}

	return &ret, nil
}
