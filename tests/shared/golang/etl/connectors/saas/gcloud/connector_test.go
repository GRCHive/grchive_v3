package gcloud

import (
	"github.com/onsi/gomega"
	"net/http"
	"testing"
)

func TestCreateGCloudConnector(t *testing.T) {
	g := gomega.NewGomegaWithT(t)

	client := &http.Client{}
	refProjectId := "123456"

	conn, err := CreateGCloudConnector(&EtlGCloudOptions{
		Client:    client,
		ProjectId: refProjectId,
	})
	g.Expect(err).To(gomega.BeNil())
	g.Expect(conn).NotTo(gomega.BeNil())
	g.Expect(conn.opts).NotTo(gomega.BeNil())
	g.Expect(conn.opts.Client).To(gomega.Equal(client))
	g.Expect(conn.opts.ProjectId).To(gomega.Equal(refProjectId))

	g.Expect(conn.users).NotTo(gomega.BeNil())
	g.Expect(conn.users.opts).To(gomega.Equal(conn.opts))
	g.Expect(conn.GetUserInterface()).To(gomega.Equal(conn.users))
}
