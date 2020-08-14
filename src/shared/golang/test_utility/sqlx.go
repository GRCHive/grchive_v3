package test_utility

import (
	"errors"
	"github.com/jmoiron/sqlx"
)

type FakeSqlx struct{}

var FakeError = errors.New("FAKE ERROR")

func (s *FakeSqlx) Queryx(string, ...interface{}) (*sqlx.Rows, error) {
	return nil, FakeError
}

func (s *FakeSqlx) MustBegin() *sqlx.Tx {
	return nil
}
