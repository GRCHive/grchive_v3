package cloudflare

import (
	"fmt"
	"github.com/onsi/gomega"
	"gitlab.com/grchive/grchive-v3/shared/etl/types"
	"gitlab.com/grchive/grchive-v3/shared/test_utility"
	"gitlab.com/grchive/grchive-v3/tests/shared/etl/connectors/saas/cloudflare_utility"
	"io/ioutil"
	"net/http"
	"strings"
	"testing"
)

func TestUserListingParse(t *testing.T) {
	g := gomega.NewGomegaWithT(t)

	client := &cloudflare_utility.MockCloudflareClient{
		AccountMembers: func() (*http.Response, error) {
			data := fmt.Sprintf(`
{"result":[{"id":"4f40b6057e64fb5cbcaee5762c324c86","user":{"id":"3acc33b57bfe3699f622ae416de1b5fa","first_name":null,"last_name":null,"email":"mike@grchive.com","two_factor_authentication_enabled":false},"status":"accepted","roles":[{"id":"33666b9c79b9a5273fc7344ff42f953d","name":"Super Administrator - All Privileges","description":"Can edit any Cloudflare setting, make purchases, update billing, and manage memberships. Super Administrators can revoke the access of other Super Administrators.","permissions":{"organization":{"read":true,"edit":true},"zone":{"read":true,"edit":false},"ssl":{"read":false,"edit":true}, "waf":{"read": false, "edit": false}}}]}],"result_info":{"page":1,"per_page":50,"total_pages":1,"count":1,"total_count":1},"success":true,"errors":[],"messages":[]}
		`)
			body := ioutil.NopCloser(strings.NewReader(data))
			return &http.Response{
				StatusCode: http.StatusOK,
				Body:       body,
			}, nil
		},
	}
	conn, err := CreateCloudflareConnector(&EtlCloudflareOptions{
		Client:    client,
		AccountId: "test",
	})
	g.Expect(err).To(gomega.BeNil())

	itf, err := conn.GetUserInterface()
	g.Expect(err).To(gomega.BeNil())

	users, source, err := itf.GetUserListing()
	g.Expect(err).To(gomega.BeNil())

	g.Expect(source).NotTo(gomega.BeNil())
	g.Expect(len(source.Commands)).To(gomega.Equal(1))

	refUsers := map[string]*types.EtlUser{
		"mike@grchive.com": &types.EtlUser{
			Username: "mike@grchive.com",
			Email:    "mike@grchive.com",
			FullName: "",
			Roles: map[string]*types.EtlRole{
				"Super Administrator - All Privileges": &types.EtlRole{
					Name: "Super Administrator - All Privileges",
					Permissions: map[string][]string{
						"organization": []string{"Read", "Edit"},
						"zone":         []string{"Read"},
						"ssl":          []string{"Edit"},
						"waf":          []string{},
					},
				},
			},
		},
	}
	test_utility.CompareUserListing(g, users, refUsers, test_utility.CompareUserListingOptions{})
}
