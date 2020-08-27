package okta

import (
	"fmt"
	"github.com/onsi/gomega"
	"gitlab.com/grchive/grchive-v3/shared/etl/types"
	"gitlab.com/grchive/grchive-v3/shared/test_utility"
	"gitlab.com/grchive/grchive-v3/tests/shared/etl/connectors/iam/okta_utility"
	"net/http"
	"testing"
	"time"
)

var refTime1 = time.Date(2010, 12, 1, 2, 3, 4, 0, time.UTC)
var refTime2 = time.Date(2012, 3, 2, 3, 4, 5, 0, time.UTC)
var refTime3 = time.Date(1999, 4, 3, 4, 5, 5, 0, time.UTC)
var refTime4 = time.Date(1990, 5, 4, 5, 6, 6, 0, time.UTC)

func createOktaClient() *okta_utility.MockOktaClient {
	return &okta_utility.MockOktaClient{
		Users: func() (*http.Response, error) {
			return test_utility.WrapHttpResponse(fmt.Sprintf(`
[{"id":"00u1akz0l37tZUMjI4x6","status":"ACTIVE","created":"%s","activated":null,"statusChanged":"2020-02-05T03:09:28.000Z","lastLogin":"2020-08-27T15:21:52.000Z","lastUpdated":"%s","passwordChanged":"2020-02-05T03:09:28.000Z","type":{"id":"oty1akyy36VmHp3Ep4x6"},"profile":{"firstName":"Michael","lastName":"Bao","mobilePhone":null,"secondEmail":null,"login":"mike@grchive.com","email":"mike@grchive.com"},"credentials":{"password":{},"emails":[{"value":"mike@grchive.com","status":"VERIFIED","type":"PRIMARY"}],"recovery_question":{"question":"What was your dream job as a child?"},"provider":{"type":"OKTA","name":"OKTA"}},"_links":{"self":{"href":"https://dev-798696.okta.com/api/v1/users/00u1akz0l37tZUMjI4x6"}}},{"id":"00u247n9hTdTgpzGB4x6","status":"ACTIVE","created":"%s","activated":"2020-02-07T03:53:04.000Z","statusChanged":"2020-02-07T03:53:04.000Z","lastLogin":"2020-07-17T18:12:44.000Z","lastUpdated":"%s","passwordChanged":null,"type":{"id":"oty1akyy36VmHp3Ep4x6"},"profile":{"firstName":"Derek","lastName":"Chin","mobilePhone":null,"secondEmail":null,"login":"derek@grchive.com","email":"derek@grchive.com"},"credentials":{"emails":[{"value":"derek@grchive.com","status":"VERIFIED","type":"PRIMARY"}],"provider":{"type":"FEDERATION","name":"FEDERATION"}},"_links":{"self":{"href":"https://dev-798696.okta.com/api/v1/users/00u247n9hTdTgpzGB4x6"}}}]
`,
				refTime1.Format(time.RFC3339),
				refTime2.Format(time.RFC3339),
				refTime3.Format(time.RFC3339),
				refTime4.Format(time.RFC3339),
			)), nil
		},
		UserRoles: map[string]okta_utility.MockOktaFn{
			"00u1akz0l37tZUMjI4x6": func() (*http.Response, error) {
				return test_utility.WrapHttpResponse(`
[{"id":"ra11akz0nxi08Ld4O4x6","label":"Super Organization Administrator","type":"SUPER_ADMIN","status":"ACTIVE","created":"2020-02-05T03:08:03.000Z","lastUpdated":"2020-02-05T03:08:03.000Z","assignmentType":"USER","_links":{"assignee":{"href":"https://dev-798696.okta.com/api/v1/users/00u1akz0l37tZUMjI4x6"}}}]
	`), nil

			},
			"00u247n9hTdTgpzGB4x6": func() (*http.Response, error) {
				return test_utility.WrapHttpResponse(`
[]
	`), nil
			},
		},
	}
}

func createConnector(g *gomega.GomegaWithT) *EtlOktaConnector {
	conn, err := CreateOktaConnector(&EtlOktaOptions{
		Client: createOktaClient(),
		Domain: "test",
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
	g.Expect(len(source.Commands)).To(gomega.Equal(3))

	refUsers := map[string]*types.EtlUser{
		"mike@grchive.com": &types.EtlUser{
			Username:       "mike@grchive.com",
			FullName:       "Michael Bao",
			Email:          "mike@grchive.com",
			CreatedTime:    &refTime1,
			LastChangeTime: &refTime2,
			Roles: map[string]*types.EtlRole{
				"Super Organization Administrator": &types.EtlRole{
					Name: "Super Organization Administrator",
				},
			},
		},
		"derek@grchive.com": &types.EtlUser{
			Username:       "derek@grchive.com",
			FullName:       "Derek Chin",
			Email:          "derek@grchive.com",
			CreatedTime:    &refTime3,
			LastChangeTime: &refTime4,
			Roles:          map[string]*types.EtlRole{},
		},
	}
	test_utility.CompareUserListing(g, users, refUsers, test_utility.CompareUserListingOptions{})
}
