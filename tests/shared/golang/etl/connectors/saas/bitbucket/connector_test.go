package bitbucket

import (
	"github.com/onsi/gomega"
	"net/http"
	"testing"
)

func TestCreateBitbucketConnector(t *testing.T) {
	g := gomega.NewGomegaWithT(t)

	client := &http.Client{}
	refGroupId := "123456"

	conn, err := CreateBitbucketConnector(&EtlBitbucketOptions{
		Client:      client,
		WorkspaceId: refGroupId,
	})
	g.Expect(err).To(gomega.BeNil())
	g.Expect(conn).NotTo(gomega.BeNil())
	g.Expect(conn.opts).NotTo(gomega.BeNil())
	g.Expect(conn.opts.Client).To(gomega.Equal(client))
	g.Expect(conn.opts.WorkspaceId).To(gomega.Equal(refGroupId))

	g.Expect(conn.users).NotTo(gomega.BeNil())
	g.Expect(conn.users.opts).To(gomega.Equal(conn.opts))
	g.Expect(conn.GetUserInterface()).To(gomega.Equal(conn.users))
}
