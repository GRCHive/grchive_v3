package heroku

import (
	"github.com/onsi/gomega"
	"net/http"
	"testing"
)

func TestCreateHerokuConnector(t *testing.T) {
	g := gomega.NewGomegaWithT(t)

	client := &http.Client{}
	refTeamName := "123456"

	conn, err := CreateHerokuConnector(&EtlHerokuOptions{
		Client:   client,
		TeamName: refTeamName,
	})
	g.Expect(err).To(gomega.BeNil())
	g.Expect(conn).NotTo(gomega.BeNil())
	g.Expect(conn.opts).NotTo(gomega.BeNil())
	g.Expect(conn.opts.Client).To(gomega.Equal(client))
	g.Expect(conn.opts.TeamName).To(gomega.Equal(refTeamName))

	g.Expect(conn.users).NotTo(gomega.BeNil())
	g.Expect(conn.users.opts).To(gomega.Equal(conn.opts))
	g.Expect(conn.GetUserInterface()).To(gomega.Equal(conn.users))
}
