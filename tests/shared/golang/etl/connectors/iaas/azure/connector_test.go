package azure

import (
	"github.com/onsi/gomega"
	"net/http"
	"testing"
)

func TestCreateAzureConnector(t *testing.T) {
	g := gomega.NewGomegaWithT(t)

	client1 := &http.Client{}
	client2 := &http.Client{}
	refSubscriptionId := "123456"

	conn, err := CreateAzureConnector(&EtlAzureOptions{
		ManagementClient: client1,
		GraphClient:      client2,
		SubscriptionId:   refSubscriptionId,
	})
	g.Expect(err).To(gomega.BeNil())
	g.Expect(conn).NotTo(gomega.BeNil())
	g.Expect(conn.opts).NotTo(gomega.BeNil())
	g.Expect(conn.opts.ManagementClient).To(gomega.Equal(client1))
	g.Expect(conn.opts.GraphClient).To(gomega.Equal(client2))
	g.Expect(conn.opts.SubscriptionId).To(gomega.Equal(refSubscriptionId))

	g.Expect(conn.users).NotTo(gomega.BeNil())
	g.Expect(conn.users.opts).To(gomega.Equal(conn.opts))
	g.Expect(conn.GetUserInterface()).To(gomega.Equal(conn.users))
}
