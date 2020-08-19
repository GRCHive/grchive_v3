package gitlab

import (
	"fmt"
	"github.com/onsi/gomega"
	"gitlab.com/grchive/grchive-v3/shared/etl/types"
	"gitlab.com/grchive/grchive-v3/shared/test_utility"
	"gitlab.com/grchive/grchive-v3/tests/shared/etl/connectors/saas/gitlab_utility"
	"io/ioutil"
	"net/http"
	"strings"
	"testing"
)

func TestUserListingParse(t *testing.T) {
	g := gomega.NewGomegaWithT(t)

	client := &gitlab_utility.MockGitlabClient{
		GroupMembersAll: func() (*http.Response, error) {
			data := fmt.Sprintf(`
[{"id":5335834,"name":"Michael Bao","username":"mbao","state":"active","avatar_url":"https://secure.gravatar.com/avatar/0499362d4cfe9943a1f0bb58005960d0?s=80\u0026d=identicon","web_url":"https://gitlab.com/mbao","access_level":50,"expires_at":null}]
		`)
			body := ioutil.NopCloser(strings.NewReader(data))
			return &http.Response{
				StatusCode: http.StatusOK,
				Body:       body,
			}, nil
		},
	}
	conn, err := CreateGitlabConnector(&EtlGitlabOptions{
		Client:  client,
		GroupId: "test",
	})
	g.Expect(err).To(gomega.BeNil())

	itf, err := conn.GetUserInterface()
	g.Expect(err).To(gomega.BeNil())

	users, source, err := itf.GetUserListing()
	g.Expect(err).To(gomega.BeNil())

	g.Expect(source).NotTo(gomega.BeNil())
	g.Expect(len(source.Commands)).To(gomega.Equal(1))

	refUsers := map[string]*types.EtlUser{
		"mbao": &types.EtlUser{
			Username: "mbao",
			FullName: "Michael Bao",
			Roles: map[string]*types.EtlRole{
				"Owner": &types.EtlRole{
					Name: "Owner",
				},
			},
		},
	}

	test_utility.CompareUserListing(g, users, refUsers, test_utility.CompareUserListingOptions{})
}
