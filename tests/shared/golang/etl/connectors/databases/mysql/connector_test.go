package mysql

import (
	"github.com/onsi/gomega"
	"gitlab.com/grchive/grchive-v3/shared/etl/connectors/databases/mysql/v5"
	"gitlab.com/grchive/grchive-v3/shared/etl/connectors/databases/mysql/v8"
	"gitlab.com/grchive/grchive-v3/shared/test_utility"
	"testing"
)

func TestCreateMysqlConnectorv8(t *testing.T) {
	g := gomega.NewGomegaWithT(t)
	db := &test_utility.FakeSqlx{}
	conn, err := CreateMysqlConnector(db, MysqlVersion{
		MajorVersion: 8,
	})
	g.Expect(err).To(gomega.BeNil())
	g.Expect(conn).NotTo(gomega.BeNil())
	g.Expect(conn.db).To(gomega.Equal(db))

	_, ok := conn.users.(*v8.EtlMysqlV8ConnectorUser)
	g.Expect(ok).To(gomega.BeTrue())

	g.Expect(conn.users).NotTo(gomega.BeNil())
	g.Expect(conn.GetUserInterface()).To(gomega.Equal(conn.users))
}

func TestCreateMysqlConnectorv5(t *testing.T) {
	g := gomega.NewGomegaWithT(t)
	db := &test_utility.FakeSqlx{}
	conn, err := CreateMysqlConnector(db, MysqlVersion{
		MajorVersion: 5,
	})
	g.Expect(err).To(gomega.BeNil())
	g.Expect(conn).NotTo(gomega.BeNil())
	g.Expect(conn.db).To(gomega.Equal(db))

	_, ok := conn.users.(*v5.EtlMysqlV5ConnectorUser)
	g.Expect(ok).To(gomega.BeTrue())

	g.Expect(conn.users).NotTo(gomega.BeNil())
	g.Expect(conn.GetUserInterface()).To(gomega.Equal(conn.users))

}
