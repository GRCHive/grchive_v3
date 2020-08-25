package linode

import (
	"github.com/onsi/gomega"
	"gitlab.com/grchive/grchive-v3/shared/etl/types"
	"gitlab.com/grchive/grchive-v3/shared/test_utility"
	"gitlab.com/grchive/grchive-v3/tests/shared/etl/connectors/iaas/linode_utility"
	"net/http"
	"testing"
)

func createLinodeClient() *linode_utility.MockLinodeClient {
	return &linode_utility.MockLinodeClient{
		AccountUsers: func() (*http.Response, error) {
			return test_utility.WrapHttpResponse(`
{"data": [{"username": "mike-grchive", "email": "mike@grchive.com", "restricted": false, "ssh_keys": [], "tfa_enabled": false}, {"username": "mike-test", "email": "mike+test@grchive.com", "restricted": true, "ssh_keys": [], "tfa_enabled": false}], "page": 1, "pages": 1, "results": 2}
`), nil
		},
		UserGrants: map[string]linode_utility.MockLinodeFn{
			"mike-test": func() (*http.Response, error) {
				return test_utility.WrapHttpResponse(`
{"linode": [], "nodebalancer": [], "domain": [], "stackscript": [], "longview": [], "image": [], "volume": [], "global": {"add_domains": false, "add_linodes": true, "add_longview": false, "longview_subscription": false, "add_stackscripts": true, "add_nodebalancers": false, "add_images": false, "add_volumes": false, "account_access": "read_only", "cancel_account": false}}
`), nil
			},
		},
	}
}

func createConnector(g *gomega.GomegaWithT) *EtlLinodeConnector {
	conn, err := CreateLinodeConnector(&EtlLinodeOptions{
		Client: createLinodeClient(),
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
		"mike-grchive": &types.EtlUser{
			Username: "mike-grchive",
			Email:    "mike@grchive.com",
			Roles:    map[string]*types.EtlRole{},
		},

		"mike-test": &types.EtlUser{
			Username: "mike-test",
			Email:    "mike+test@grchive.com",
			Roles: map[string]*types.EtlRole{
				"Global": &types.EtlRole{
					Name: "Global",
					Permissions: map[string][]string{
						"Grants": []string{
							"AddLinodes",
							"AddStackScripts",
						},
						"AccountAccess": []string{"read_only"},
					},
				},
				"Linode":       &types.EtlRole{Name: "Linode"},
				"Domain":       &types.EtlRole{Name: "Domain"},
				"NodeBalancer": &types.EtlRole{Name: "NodeBalancer"},
				"Image":        &types.EtlRole{Name: "Image"},
				"LongView":     &types.EtlRole{Name: "LongView"},
				"StackScript":  &types.EtlRole{Name: "StackScript"},
				"Volume":       &types.EtlRole{Name: "Volume"},
			},
		},
	}
	test_utility.CompareUserListing(g, users, refUsers, test_utility.CompareUserListingOptions{})
}
