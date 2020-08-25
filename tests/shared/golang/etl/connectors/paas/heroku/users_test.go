package heroku

import (
	"fmt"
	"github.com/onsi/gomega"
	"gitlab.com/grchive/grchive-v3/shared/etl/types"
	"gitlab.com/grchive/grchive-v3/shared/test_utility"
	"gitlab.com/grchive/grchive-v3/tests/shared/etl/connectors/paas/heroku_utility"
	"net/http"
	"testing"
	"time"
)

var refTime1 = time.Date(2012, 10, 12, 3, 5, 6, 0, time.UTC)

func TestGetUserListing(t *testing.T) {
	g := gomega.NewGomegaWithT(t)

	client := &heroku_utility.MockHerokuClient{
		TeamMembers: func() (*http.Response, error) {
			return test_utility.WrapHttpResponse(fmt.Sprintf(`
[{"id":"a15ac3ce-fad5-4ea8-8e5e-7b352e6b3c28","created_at":"%s","email":"mike@grchive.com","federated":false,"identity_provider":null,"role":"admin","updated_at":"2020-08-25T19:17:20Z","user":{"id":"c6943e34-95ec-4d6c-87bd-90bbeed025dc","email":"mike@grchive.com","name":"Michael Bao"}}]
`, refTime1.Format(time.RFC3339))), nil
		},
	}
	conn, err := CreateHerokuConnector(&EtlHerokuOptions{
		Client:   client,
		TeamName: "test",
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
			Username:    "mike@grchive.com",
			Email:       "mike@grchive.com",
			FullName:    "Michael Bao",
			CreatedTime: &refTime1,
			Roles: map[string]*types.EtlRole{
				"admin": &types.EtlRole{
					Name: "admin",
				},
			},
		},
	}

	test_utility.CompareUserListing(g, users, refUsers, test_utility.CompareUserListingOptions{})
}
