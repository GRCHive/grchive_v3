package mariadb

import (
	"gitlab.com/grchive/grchive-v3/shared/etl/connectors"
	"gitlab.com/grchive/grchive-v3/shared/etl/connectors/databases"
)

type EtlMariadbConnector struct {
	db    databases.SqlxLike
	users connectors.EtlConnectorUserInterface
}

func (c *EtlMariadbConnector) GetUserInterface() (connectors.EtlConnectorUserInterface, error) {
	return c.users, nil
}

func CreateMariadbConnector(db databases.SqlxLike) (*EtlMariadbConnector, error) {
	var err error
	ret := EtlMariadbConnector{
		db: db,
	}

	ret.users = &EtlMariadbConnectorUser{
		db: &databases.DB{
			SqlxLike: db,
		},
	}
	if err != nil {
		return nil, err
	}

	return &ret, nil
}
