package mssql

import (
	"github.com/onsi/gomega"
	"gitlab.com/grchive/grchive-v3/shared/test_utility"
	"testing"
)

func TestCreateMssqlConnector(t *testing.T) {
	g := gomega.NewGomegaWithT(t)
	db := &test_utility.FakeSqlx{}
	conn, err := CreateMssqlConnector(db)
	g.Expect(err).To(gomega.BeNil())
	g.Expect(conn).NotTo(gomega.BeNil())
	g.Expect(conn.db).To(gomega.Equal(db))

	g.Expect(conn.users).NotTo(gomega.BeNil())
	g.Expect(conn.users.db.SqlxLike).To(gomega.Equal(db))
	g.Expect(conn.GetUserInterface()).To(gomega.Equal(conn.users))
}
