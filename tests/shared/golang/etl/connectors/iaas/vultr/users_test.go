package vultr

import (
	"github.com/onsi/gomega"
	"gitlab.com/grchive/grchive-v3/shared/etl/types"
	"gitlab.com/grchive/grchive-v3/shared/test_utility"
	"gitlab.com/grchive/grchive-v3/tests/shared/etl/connectors/iaas/vultr_utility"
	"net/http"
	"testing"
)

func createVultrClient() *vultr_utility.MockVultrClient {
	return &vultr_utility.MockVultrClient{
		GetUsers: func() (*http.Response, error) {
			return test_utility.WrapHttpResponse(`
{"users":[{"id":"dev-preview-mjrwcnlggq2tenrwmm2gi","name":"Mike Test","email":"mike+test@grchive.com","api_enabled":"yes","acls":["firewall"]}],"meta":{"total":1,"links":{"next":"","prev":""}}}
`), nil
		},
		GetAccountInfo: func() (*http.Response, error) {
			return test_utility.WrapHttpResponse(`
{"account":{"balance":0,"pending_charges":0,"name":"Michael Bao","email":"mike@grchive.com","acls":["manage_users","subscriptions_view","subscriptions","billing","support","provisioning","dns","abuse","upgrade","firewall","alerts","objstore","loadbalancer"]}}
`), nil
		},
	}
}

func createConnector(g *gomega.GomegaWithT) *EtlVultrConnector {
	conn, err := CreateVultrConnector(&EtlVultrOptions{
		Client: createVultrClient(),
	})
	g.Expect(err).To(gomega.BeNil())
	return conn
}

func TestGetUserListing(t *testing.T) {
	g := gomega.NewGomegaWithT(t)
	conn := createConnector(g)

	itf, err := conn.GetUserInterface()
	g.Expect(err).To(gomega.BeNil())

	users, source, err := itf.GetUserListing()
	g.Expect(err).To(gomega.BeNil())

	g.Expect(err).To(gomega.BeNil())
	g.Expect(len(source.Commands)).To(gomega.Equal(2))

	refUsers := map[string]*types.EtlUser{
		"mike+test@grchive.com": &types.EtlUser{
			Username: "mike+test@grchive.com",
			FullName: "Mike Test",
			Email:    "mike+test@grchive.com",
			Roles: map[string]*types.EtlRole{
				"Self": &types.EtlRole{
					Name: "Self",
					Permissions: map[string][]string{
						"Self": []string{
							"firewall",
						},
					},
				},
			},
		},

		"mike@grchive.com": &types.EtlUser{
			Username: "mike@grchive.com",
			FullName: "Michael Bao",
			Email:    "mike@grchive.com",
			Roles: map[string]*types.EtlRole{
				"Self": &types.EtlRole{
					Name: "Self",
					Permissions: map[string][]string{
						"Self": []string{
							"manage_users",
							"subscriptions_view",
							"subscriptions",
							"billing",
							"support",
							"provisioning",
							"dns",
							"abuse",
							"upgrade",
							"firewall",
							"alerts",
							"objstore",
							"loadbalancer",
						},
					},
				},
			},
		},
	}
	test_utility.CompareUserListing(g, users, refUsers, test_utility.CompareUserListingOptions{})

}
