package ibm

import (
	"gitlab.com/grchive/grchive-v3/shared/etl/connectors"
	"gitlab.com/grchive/grchive-v3/shared/etl/connectors/databases"
)

type EtlIBMConnector struct {
	db    databases.SqlxLike
	users *EtlIBMConnectorUser
}

func (c *EtlIBMConnector) GetUserInterface() (connectors.EtlConnectorUserInterface, error) {
	return c.users, nil
}

func CreateIBMConnector(db databases.SqlxLike) (*EtlIBMConnector, error) {
	var err error
	ret := EtlIBMConnector{
		db: db,
	}
	ret.users, err = createIBMConnectorUser(&databases.DB{
		SqlxLike: db,
	})
	if err != nil {
		return nil, err
	}

	return &ret, nil
}
