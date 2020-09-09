package oracle

import (
	"gitlab.com/grchive/grchive-v3/shared/etl/connectors"
	"gitlab.com/grchive/grchive-v3/shared/etl/connectors/databases"
)

type EtlOracleConnector struct {
	db    databases.SqlxLike
	users *EtlOracleConnectorUser
}

func (c *EtlOracleConnector) GetUserInterface() (connectors.EtlConnectorUserInterface, error) {
	return c.users, nil
}

func CreateOracleConnector(db databases.SqlxLike) (*EtlOracleConnector, error) {
	var err error
	ret := EtlOracleConnector{
		db: db,
	}
	ret.users, err = createOracleConnectorUser(&databases.DB{
		SqlxLike: db,
	})
	if err != nil {
		return nil, err
	}

	return &ret, nil
}
