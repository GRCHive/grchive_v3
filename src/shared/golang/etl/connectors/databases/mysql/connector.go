package mysql

import (
	"errors"
	"gitlab.com/grchive/grchive-v3/shared/etl/connectors"
	"gitlab.com/grchive/grchive-v3/shared/etl/connectors/databases"
	"gitlab.com/grchive/grchive-v3/shared/etl/connectors/databases/mysql/v5"
	"gitlab.com/grchive/grchive-v3/shared/etl/connectors/databases/mysql/v8"
	"strconv"
	"strings"
)

type EtlMysqlInterfaceFactory interface {
	CreateUserInterface(*databases.DB) (connectors.EtlConnectorUserInterface, error)
}

type EtlMysqlConnector struct {
	db    databases.SqlxLike
	users connectors.EtlConnectorUserInterface
}

type MysqlVersion struct {
	MajorVersion int
	MinorVersion int
	AuxVersion   string
}

func (c *EtlMysqlConnector) GetUserInterface() (connectors.EtlConnectorUserInterface, error) {
	return c.users, nil
}

func ObtainMysqlVersion(db databases.SqlxLike) (*MysqlVersion, error) {
	rows, err := db.Queryx("SELECT VERSION()")
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	rows.Next()

	version := ""
	err = rows.Scan(&version)
	if err != nil {
		return nil, err
	}

	splitVersion := strings.Split(version, ".")

	majorVersion, err := strconv.Atoi(splitVersion[0])
	if err != nil {
		return nil, err
	}

	minorVersion, err := strconv.Atoi(splitVersion[1])
	if err != nil {
		return nil, err
	}

	return &MysqlVersion{
		MajorVersion: majorVersion,
		MinorVersion: minorVersion,
		AuxVersion:   splitVersion[2],
	}, nil
}

func CreateMysqlConnector(db databases.SqlxLike, version MysqlVersion) (*EtlMysqlConnector, error) {
	var err error
	ret := EtlMysqlConnector{
		db: db,
	}

	var versionFactory EtlMysqlInterfaceFactory
	if err != nil {
		return nil, err
	}

	if version.MajorVersion == 8 {
		versionFactory = &v8.InterfaceFactory{}
	} else if version.MajorVersion == 5 {
		versionFactory = &v5.InterfaceFactory{}
	} else {
		return nil, errors.New("Unsupported MySQL version.")
	}

	ret.users, err = versionFactory.CreateUserInterface(&databases.DB{
		SqlxLike: db,
	})
	if err != nil {
		return nil, err
	}

	return &ret, nil
}
