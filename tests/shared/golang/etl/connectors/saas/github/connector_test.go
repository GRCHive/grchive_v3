package github

import (
	"github.com/onsi/gomega"
	"net/http"
	"testing"
)

func TestCreateGithubConnector(t *testing.T) {
	g := gomega.NewGomegaWithT(t)

	client := &http.Client{}
	refOrgId := "123456"

	conn, err := CreateGithubConnector(&EtlGithubOptions{
		Client: client,
		OrgId:  refOrgId,
	})
	g.Expect(err).To(gomega.BeNil())
	g.Expect(conn).NotTo(gomega.BeNil())
	g.Expect(conn.opts).NotTo(gomega.BeNil())
	g.Expect(conn.opts.Client).To(gomega.Equal(client))
	g.Expect(conn.opts.OrgId).To(gomega.Equal(refOrgId))

	g.Expect(conn.users).NotTo(gomega.BeNil())
	g.Expect(conn.users.opts).To(gomega.Equal(conn.opts))
	g.Expect(conn.GetUserInterface()).To(gomega.Equal(conn.users))
}
