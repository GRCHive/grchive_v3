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
	"time"
)

func TestUserListingParse(t *testing.T) {
	g := gomega.NewGomegaWithT(t)

	u1Time := time.Date(2000, 12, 10, 12, 23, 43, 500, time.UTC)
	u2Time := time.Date(2006, 1, 5, 3, 10, 33, 100, time.UTC)

	client := &github_utility.MockGithubClient{
		GraphQL: func() (*http.Response, error) {
			data := fmt.Sprintf(`
{"data":{"organization":{"name":"GRCHive","membersWithRole":{"edges":[{"node":{"name":"Michael Bao","login":"b3h47pte","createdAt":"%s"},"role":"MEMBER"},{"node":{"name":null,"login":"mikebao-grchive","createdAt":"%s"},"role":"ADMIN"}],"pageInfo":{"endCursor":"Y3Vyc29yOnYyOpHOBClgEA==","hasNextPage":false}}}}}
		`, u1Time.Format(time.RFC3339), u2Time.Format(time.RFC3339))
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
	g.Expect(len(source.Commands)).To(gomega.Equal(1))

	refUsers := map[string]*types.EtlUser{
		"mikebao-grchive": &types.EtlUser{
			Username:    "mikebao-grchive",
			FullName:    "",
			CreatedTime: &u2Time,
			Roles: map[string]*types.EtlRole{
				"ADMIN": &types.EtlRole{
					Name: "ADMIN",
				},
			},
		},
		"b3h47pte": &types.EtlUser{
			Username:    "b3h47pte",
			FullName:    "Michael Bao",
			CreatedTime: &u1Time,
			Roles: map[string]*types.EtlRole{
				"MEMBER": &types.EtlRole{
					Name: "MEMBER",
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
		g.Expect(u.CreatedTime).NotTo(gomega.BeNil())
		g.Expect(*u.CreatedTime).To(gomega.BeTemporally("~", *refU.CreatedTime, time.Second))
		g.Expect(len(u.Roles)).To(gomega.Equal(len(refU.Roles)))

		for _, r := range u.Roles {
			refRole, ok := refU.Roles[r.Name]
			g.Expect(ok).To(gomega.BeTrue(), "Finding role: "+r.Name)
			g.Expect(r.Name).To(gomega.Equal(refRole.Name))
			g.Expect(len(r.Permissions)).To(gomega.Equal(0))
		}
	}
}
