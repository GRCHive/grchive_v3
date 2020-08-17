package github

import (
	"fmt"
	"github.com/onsi/gomega"
	"gitlab.com/grchive/grchive-v3/shared/etl/types"
	"gitlab.com/grchive/grchive-v3/tests/shared/etl/connectors/saas/github_utility"
	"io/ioutil"
	"net/http"
	"strings"
	"testing"
)

func TestUserListingParse(t *testing.T) {
	g := gomega.NewGomegaWithT(t)

	client := &github_utility.MockGithubClient{
		OrgAdminsList: func() (*http.Response, error) {
			data := fmt.Sprintf(`
[{"login":"mikebao-grchive","id":69820432,"node_id":"MDQ6VXNlcjY5ODIwNDMy","avatar_url":"https://avatars2.githubusercontent.com/u/69820432?v=4","gravatar_id":"","url":"https://api.github.com/users/mikebao-grchive","html_url":"https://github.com/mikebao-grchive","followers_url":"https://api.github.com/users/mikebao-grchive/followers","following_url":"https://api.github.com/users/mikebao-grchive/following{/other_user}","gists_url":"https://api.github.com/users/mikebao-grchive/gists{/gist_id}","starred_url":"https://api.github.com/users/mikebao-grchive/starred{/owner}{/repo}","subscriptions_url":"https://api.github.com/users/mikebao-grchive/subscriptions","organizations_url":"https://api.github.com/users/mikebao-grchive/orgs","repos_url":"https://api.github.com/users/mikebao-grchive/repos","events_url":"https://api.github.com/users/mikebao-grchive/events{/privacy}","received_events_url":"https://api.github.com/users/mikebao-grchive/received_events","type":"User","site_admin":false}]
		`)
			body := ioutil.NopCloser(strings.NewReader(data))
			return &http.Response{
				StatusCode: http.StatusOK,
				Body:       body,
			}, nil
		},
		OrgMembersList: func() (*http.Response, error) {
			data := fmt.Sprintf(`
[{"login":"b3h47pte","id":1286844,"node_id":"MDQ6VXNlcjEyODY4NDQ=","avatar_url":"https://avatars0.githubusercontent.com/u/1286844?v=4","gravatar_id":"","url":"https://api.github.com/users/b3h47pte","html_url":"https://github.com/b3h47pte","followers_url":"https://api.github.com/users/b3h47pte/followers","following_url":"https://api.github.com/users/b3h47pte/following{/other_user}","gists_url":"https://api.github.com/users/b3h47pte/gists{/gist_id}","starred_url":"https://api.github.com/users/b3h47pte/starred{/owner}{/repo}","subscriptions_url":"https://api.github.com/users/b3h47pte/subscriptions","organizations_url":"https://api.github.com/users/b3h47pte/orgs","repos_url":"https://api.github.com/users/b3h47pte/repos","events_url":"https://api.github.com/users/b3h47pte/events{/privacy}","received_events_url":"https://api.github.com/users/b3h47pte/received_events","type":"User","site_admin":false}]
		`)
			body := ioutil.NopCloser(strings.NewReader(data))
			return &http.Response{
				StatusCode: http.StatusOK,
				Body:       body,
			}, nil
		},
	}
	conn, err := CreateGithubConnector(&EtlGithubOptions{
		Client: client,
		OrgId:  "test",
	})
	g.Expect(err).To(gomega.BeNil())

	itf, err := conn.GetUserInterface()
	g.Expect(err).To(gomega.BeNil())

	users, source, err := itf.GetUserListing()
	g.Expect(err).To(gomega.BeNil())

	g.Expect(source).NotTo(gomega.BeNil())
	g.Expect(len(source.Commands)).To(gomega.Equal(2))

	refUsers := map[string]*types.EtlUser{
		"mikebao-grchive": &types.EtlUser{
			Username: "mikebao-grchive",
			Roles: map[string]*types.EtlRole{
				"admin": &types.EtlRole{
					Name: "admin",
				},
			},
		},
		"b3h47pte": &types.EtlUser{
			Username: "b3h47pte",
			Roles: map[string]*types.EtlRole{
				"member": &types.EtlRole{
					Name: "member",
				},
			},
		},
	}

	g.Expect(len(users)).To(gomega.Equal(len(refUsers)))
	for _, u := range users {
		refU, ok := refUsers[u.Username]
		g.Expect(ok).To(gomega.BeTrue(), "Finding username: "+u.Username)

		g.Expect(u.Username).To(gomega.Equal(refU.Username))
		g.Expect(len(u.Roles)).To(gomega.Equal(len(refU.Roles)))

		for _, r := range u.Roles {
			refRole, ok := refU.Roles[r.Name]
			g.Expect(ok).To(gomega.BeTrue(), "Finding role: "+r.Name)
			g.Expect(r.Name).To(gomega.Equal(refRole.Name))
			g.Expect(len(r.Permissions)).To(gomega.Equal(0))
		}
	}
}
