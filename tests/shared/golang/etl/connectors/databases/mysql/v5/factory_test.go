package v5

import (
	"github.com/onsi/gomega"
	"gitlab.com/grchive/grchive-v3/shared/etl/connectors/databases"
	"gitlab.com/grchive/grchive-v3/shared/test_utility"
	"testing"
)

func TestInterfaceFactory(t *testing.T) {
	g := gomega.NewGomegaWithT(t)
	db := &test_utility.FakeSqlx{}

	dbWrap := &databases.DB{
		SqlxLike: db,
	}

	factory := InterfaceFactory{}
	user, err := factory.CreateUserInterface(dbWrap)
	g.Expect(err).To(gomega.BeNil())

	tuser, ok := user.(*EtlMysqlV5ConnectorUser)
	g.Expect(ok).To(gomega.BeTrue())
	g.Expect(tuser.db).To(gomega.Equal(dbWrap))
}
