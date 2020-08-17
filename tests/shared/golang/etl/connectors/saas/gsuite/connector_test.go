package gsuite

import (
	"github.com/onsi/gomega"
	"net/http"
	"testing"
)

func TestCreateGSuiteConnector(t *testing.T) {
	g := gomega.NewGomegaWithT(t)

	client := &http.Client{}
	refCustomerId := "123456"

	conn, err := CreateGSuiteConnector(&EtlGSuiteOptions{
		Client:     client,
		CustomerId: refCustomerId,
	})
	g.Expect(err).To(gomega.BeNil())
	g.Expect(conn).NotTo(gomega.BeNil())
	g.Expect(conn.opts).NotTo(gomega.BeNil())
	g.Expect(conn.opts.Client).To(gomega.Equal(client))
	g.Expect(conn.opts.CustomerId).To(gomega.Equal(refCustomerId))

	g.Expect(conn.users).NotTo(gomega.BeNil())
	g.Expect(conn.users.opts).To(gomega.Equal(conn.opts))
	g.Expect(conn.GetUserInterface()).To(gomega.Equal(conn.users))
}
