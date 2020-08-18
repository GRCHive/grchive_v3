package bitbucket

import (
	"fmt"
	"github.com/onsi/gomega"
	"gitlab.com/grchive/grchive-v3/shared/etl/types"
	"gitlab.com/grchive/grchive-v3/tests/shared/etl/connectors/saas/bitbucket_utility"
	"io/ioutil"
	"net/http"
	"strings"
	"testing"
)

func TestUserListingParse(t *testing.T) {
	g := gomega.NewGomegaWithT(t)

	client := &bitbucket_utility.MockBitbucketClient{
		WorkspaceMembersAll: func() (*http.Response, error) {
			data := fmt.Sprintf(`
{"pagelen": 50, "values": [{"links": {"self": {"href": "https://api.bitbucket.org/2.0/workspaces/grchive/members/%7B69f111dc-bbdf-447e-9e60-c586aae2820f%7D"}}, "permission": "owner", "last_accessed": null, "user": {"display_name": "Michael Bao", "uuid": "{69f111dc-bbdf-447e-9e60-c586aae2820f}", "links": {"self": {"href": "https://api.bitbucket.org/2.0/users/%7B69f111dc-bbdf-447e-9e60-c586aae2820f%7D"}, "html": {"href": "https://bitbucket.org/%7B69f111dc-bbdf-447e-9e60-c586aae2820f%7D/"}, "avatar": {"href": "https://secure.gravatar.com/avatar/0499362d4cfe9943a1f0bb58005960d0?d=https%3A%2F%2Favatar-management--avatars.us-west-2.prod.public.atl-paas.net%2Finitials%2FMB-2.png"}}, "nickname": "Michael Bao", "type": "user", "account_id": "5f3c1a991ac29c0045d3e93d"}, "workspace": {"slug": "grchive", "type": "workspace", "name": "grchive", "links": {"self": {"href": "https://api.bitbucket.org/2.0/workspaces/grchive"}, "html": {"href": "https://bitbucket.org/grchive/"}, "avatar": {"href": "https://bitbucket.org/workspaces/grchive/avatar/?ts=1597774568"}}, "uuid": "{44cf0298-fa55-4d9e-88f8-7b5576a8de8b}"}, "type": "workspace_membership", "added_on": null}], "page": 1, "size": 1}
		`)
			body := ioutil.NopCloser(strings.NewReader(data))
			return &http.Response{
				StatusCode: http.StatusOK,
				Body:       body,
			}, nil
		},
	}
	conn, err := CreateBitbucketConnector(&EtlBitbucketOptions{
		Client:      client,
		WorkspaceId: "test",
	})
	g.Expect(err).To(gomega.BeNil())

	itf, err := conn.GetUserInterface()
	g.Expect(err).To(gomega.BeNil())

	users, source, err := itf.GetUserListing()
	g.Expect(err).To(gomega.BeNil())

	g.Expect(source).NotTo(gomega.BeNil())
	g.Expect(len(source.Commands)).To(gomega.Equal(1))

	refUsers := map[string]*types.EtlUser{
		"5f3c1a991ac29c0045d3e93d": &types.EtlUser{
			Username: "5f3c1a991ac29c0045d3e93d",
			FullName: "Michael Bao",
			Roles: map[string]*types.EtlRole{
				"owner": &types.EtlRole{
					Name: "owner",
				},
			},
		},
	}

	g.Expect(len(users)).To(gomega.Equal(len(refUsers)))
	for _, u := range users {
		refU, ok := refUsers[u.Username]
		g.Expect(ok).To(gomega.BeTrue(), "Finding username: "+u.Username)

		g.Expect(u.Username).To(gomega.Equal(refU.Username))
		g.Expect(u.FullName).To(gomega.Equal(refU.FullName))
		g.Expect(len(u.Roles)).To(gomega.Equal(len(refU.Roles)))

		for _, r := range u.Roles {
			refRole, ok := refU.Roles[r.Name]
			g.Expect(ok).To(gomega.BeTrue(), "Finding role: "+r.Name)
			g.Expect(r.Name).To(gomega.Equal(refRole.Name))
			g.Expect(len(r.Permissions)).To(gomega.Equal(0))
		}
	}
}
