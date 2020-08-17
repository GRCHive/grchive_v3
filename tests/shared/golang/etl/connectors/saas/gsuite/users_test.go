package gsuite

import (
	"fmt"
	"github.com/onsi/gomega"
	"gitlab.com/grchive/grchive-v3/shared/etl/types"
	"gitlab.com/grchive/grchive-v3/tests/shared/etl/connectors/saas/gsuite_utility"
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

	client := &gsuite_utility.MockGSuiteClient{
		DirectoryUsersList: func() (*http.Response, error) {
			data := fmt.Sprintf(`
		{
		  "kind": "admin#directory#users",
		  "etag": "\"7zy5N0aKZHNqctZjjdvw-wIf5MQeqDz49iNO-ev10U8/AHoHAqrBxyQG1gYlawuxhlWDwgY\"",
		  "users": [
			{
			  "kind": "admin#directory#user",
			  "id": "1",
			  "primaryEmail": "derek@grchive.com",
			  "name": {
				"givenName": "Derek",
				"familyName": "Chin",
				"fullName": "Derek Chin"
			  },
			  "isAdmin": false,
			  "isDelegatedAdmin": true,
			  "creationTime": "%s"
			},
			{
			  "kind": "admin#directory#user",
			  "id": "2",
			  "primaryEmail": "mike@grchive.com",
			  "name": {
				"givenName": "Michael",
				"familyName": "Bao",
				"fullName": "Michael Bao"
			  },
			  "isAdmin": true,
			  "isDelegatedAdmin": false,
			  "creationTime": "%s"
			}
		  ]
		}
		`, u1Time.Format(time.RFC3339), u2Time.Format(time.RFC3339))
			body := ioutil.NopCloser(strings.NewReader(data))
			return &http.Response{
				StatusCode: http.StatusOK,
				Body:       body,
			}, nil

		},
	}
	conn, err := CreateGSuiteConnector(&EtlGSuiteOptions{
		Client:     client,
		CustomerId: "12345",
	})
	g.Expect(err).To(gomega.BeNil())

	itf, err := conn.GetUserInterface()
	g.Expect(err).To(gomega.BeNil())

	users, source, err := itf.GetUserListing()
	g.Expect(err).To(gomega.BeNil())

	g.Expect(source).NotTo(gomega.BeNil())
	g.Expect(len(source.Commands)).To(gomega.Equal(1))

	refUsers := map[string]*types.EtlUser{
		"derek@grchive.com": &types.EtlUser{
			Username:    "derek@grchive.com",
			FullName:    "Derek Chin",
			Email:       "derek@grchive.com",
			CreatedTime: &u1Time,
			Roles: map[string]*types.EtlRole{
				"delegatedAdmin": &types.EtlRole{
					Name: "delegatedAdmin",
				},
			},
		},
		"mike@grchive.com": &types.EtlUser{
			Username:    "mike@grchive.com",
			FullName:    "Michael Bao",
			Email:       "mike@grchive.com",
			CreatedTime: &u2Time,
			Roles: map[string]*types.EtlRole{
				"admin": &types.EtlRole{
					Name: "admin",
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
		g.Expect(u.Email).To(gomega.Equal(refU.Email))
		g.Expect(len(u.Roles)).To(gomega.Equal(len(refU.Roles)))
		g.Expect(*u.CreatedTime).To(gomega.BeTemporally("~", *refU.CreatedTime, time.Second))

		for _, r := range u.Roles {
			refRole, ok := refU.Roles[r.Name]
			g.Expect(ok).To(gomega.BeTrue(), "Finding role: "+r.Name)
			g.Expect(r.Name).To(gomega.Equal(refRole.Name))
			g.Expect(len(r.Permissions)).To(gomega.Equal(0))
		}
	}
}
