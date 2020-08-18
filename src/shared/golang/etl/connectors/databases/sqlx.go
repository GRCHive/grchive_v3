package databases

import (
	"github.com/jmoiron/sqlx"
	"gitlab.com/grchive/grchive-v3/shared/etl/connectors"
	"strconv"
)

type SqlxLike interface {
	Queryx(string, ...interface{}) (*sqlx.Rows, error)
	MustBegin() *sqlx.Tx
}

type DB struct {
	SqlxLike
}

func (d *DB) LoggedQuery(query string, args ...interface{}) (*sqlx.Rows, *connectors.EtlCommandInfo, error) {
	params := map[string]interface{}{}
	cmd := connectors.EtlCommandInfo{
		Command:    query,
		Parameters: params,
	}

	for i, val := range args {
		params[strconv.Itoa(i+1)] = val
	}

	rows, err := d.Queryx(query, args...)
	return rows, &cmd, err
}
