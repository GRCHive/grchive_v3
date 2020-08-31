package auth0

import (
	"fmt"
	"github.com/onsi/gomega"
	"gitlab.com/grchive/grchive-v3/shared/etl/types"
	"gitlab.com/grchive/grchive-v3/shared/test_utility"
	"gitlab.com/grchive/grchive-v3/tests/shared/etl/connectors/iam/auth0_utility"
	"net/http"
	"testing"
	"time"
)

var refTime1 = time.Date(2010, 12, 1, 2, 3, 4, 0, time.UTC)
var refTime2 = time.Date(2012, 3, 2, 3, 4, 5, 0, time.UTC)
var refTime3 = time.Date(1999, 4, 3, 4, 5, 5, 0, time.UTC)
var refTime4 = time.Date(1990, 5, 4, 5, 6, 6, 0, time.UTC)

func createAuth0Client() *auth0_utility.MockAuth0Client {
	return &auth0_utility.MockAuth0Client{
		Users: func() (*http.Response, error) {
			return test_utility.WrapHttpResponse(fmt.Sprintf(`
[{"created_at":"%s","email":"mike+test@grchive.com","email_verified":false,"identities":[{"user_id":"5f4d3020146161006d256bce","provider":"auth0","connection":"Username-Password-Authentication","isSocial":false}],"name":"Mike Bao","nickname":"mike+test","picture":"https://s.gravatar.com/avatar/52da8b064b5bbe23cac46323153fe0f4?s=480&r=pg&d=https%3A%2F%2Fcdn.auth0.com%2Favatars%2Fmi.png","updated_at":"2020-08-31T18:41:41.883Z","user_id":"auth0|5f4d3020146161006d256bce","last_login":"2020-08-31T17:15:27.722Z","last_ip":"96.225.71.232","logins_count":1}]
`,
				refTime1.Format(time.RFC3339),
			)), nil
		},
	}
}

func createConnector(g *gomega.GomegaWithT) *EtlAuth0Connector {
	conn, err := CreateAuth0Connector(&EtlAuth0Options{
		Client: createAuth0Client(),
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
	g.Expect(len(source.Commands)).To(gomega.Equal(1))

	refUsers := map[string]*types.EtlUser{
		"mike+test@grchive.com": &types.EtlUser{
			Username:    "mike+test@grchive.com",
			FullName:    "Mike Bao",
			Email:       "mike+test@grchive.com",
			CreatedTime: &refTime1,
			Roles:       map[string]*types.EtlRole{},
		},
	}
	test_utility.CompareUserListing(g, users, refUsers, test_utility.CompareUserListingOptions{})
}
