package mariadb

import (
	"github.com/onsi/gomega"
	"gitlab.com/grchive/grchive-v3/shared/test_utility"
	"testing"
)

func TestCreateMariadbConnectorv8(t *testing.T) {
	g := gomega.NewGomegaWithT(t)
	db := &test_utility.FakeSqlx{}
	conn, err := CreateMariadbConnector(db)
	g.Expect(err).To(gomega.BeNil())
	g.Expect(conn).NotTo(gomega.BeNil())
	g.Expect(conn.db).To(gomega.Equal(db))

	_, ok := conn.users.(*EtlMariadbConnectorUser)
	g.Expect(ok).To(gomega.BeTrue())

	g.Expect(conn.users).NotTo(gomega.BeNil())
	g.Expect(conn.GetUserInterface()).To(gomega.Equal(conn.users))
}
