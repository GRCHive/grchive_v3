package auth0

import (
	"github.com/onsi/gomega"
	"net/http"
	"testing"
)

func TestCreateAuth0Connector(t *testing.T) {
	g := gomega.NewGomegaWithT(t)

	client := &http.Client{}
	refDomain := "test"

	conn, err := CreateAuth0Connector(&EtlAuth0Options{
		Client: client,
		Domain: refDomain,
	})
	g.Expect(err).To(gomega.BeNil())
	g.Expect(conn).NotTo(gomega.BeNil())
	g.Expect(conn.opts).NotTo(gomega.BeNil())
	g.Expect(conn.opts.Client).To(gomega.Equal(client))
	g.Expect(conn.opts.Domain).To(gomega.Equal(refDomain))

	g.Expect(conn.users).NotTo(gomega.BeNil())
	g.Expect(conn.users.opts).To(gomega.Equal(conn.opts))
	g.Expect(conn.GetUserInterface()).To(gomega.Equal(conn.users))
}
