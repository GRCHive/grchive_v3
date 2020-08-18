package sqlx_test

import (
	"github.com/onsi/gomega"
	"gitlab.com/grchive/grchive-v3/shared/etl/connectors/databases"
	"gitlab.com/grchive/grchive-v3/shared/test_utility"
	"testing"
)

func TestLoggedQuery(t *testing.T) {
	g := gomega.NewGomegaWithT(t)

	db := databases.DB{
		SqlxLike: &test_utility.FakeSqlx{},
	}

	{
		fakeQuery := "TEST QUERY 1234"
		rows, cmd, err := db.LoggedQuery(fakeQuery)
		g.Expect(rows).To(gomega.BeNil())
		g.Expect(cmd.Command).To(gomega.Equal(fakeQuery))
		g.Expect(len(cmd.Parameters.(map[string]interface{}))).To(gomega.Equal(0))
		g.Expect(cmd.RawData).To(gomega.Equal(""))
		g.Expect(err).To(gomega.Equal(test_utility.FakeError))
	}

	{
		fakeQuery := "rando M$!!"
		args := []interface{}{"param0", "param1", "param2"}
		rows, cmd, err := db.LoggedQuery(fakeQuery, args...)
		g.Expect(rows).To(gomega.BeNil())
		g.Expect(cmd.Command).To(gomega.Equal(fakeQuery))
		g.Expect(len(cmd.Parameters.(map[string]interface{}))).To(gomega.Equal(len(args)))
		g.Expect(cmd.Parameters.(map[string]interface{})["1"]).To(gomega.Equal(args[0]))
		g.Expect(cmd.Parameters.(map[string]interface{})["2"]).To(gomega.Equal(args[1]))
		g.Expect(cmd.Parameters.(map[string]interface{})["3"]).To(gomega.Equal(args[2]))
		g.Expect(cmd.RawData).To(gomega.Equal(""))
		g.Expect(err).To(gomega.Equal(test_utility.FakeError))
	}
}
