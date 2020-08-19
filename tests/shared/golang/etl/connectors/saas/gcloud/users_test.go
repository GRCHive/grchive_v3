package gcloud

import (
	"fmt"
	"github.com/onsi/gomega"
	"gitlab.com/grchive/grchive-v3/shared/etl/types"
	"gitlab.com/grchive/grchive-v3/shared/test_utility"
	"gitlab.com/grchive/grchive-v3/tests/shared/etl/connectors/saas/gcloud_utility"
	"io/ioutil"
	"net/http"
	"strings"
	"testing"
)

func TestUserListingParse(t *testing.T) {
	g := gomega.NewGomegaWithT(t)

	client := &gcloud_utility.MockGCloudClient{
		GetUserListing: func() (*http.Response, error) {
			data := fmt.Sprintf(`
{
  "version": 1,
  "etag": "BwWtPPmYkyA=",
  "bindings": [
    {
      "role": "organizations/94248544035/roles/CustomRole294",
      "members": [
        "serviceAccount:grchive-service-account@grchive-v3.iam.gserviceaccount.com",
        "user:mike@grchive.com"
      ]
    },
    {
      "role": "roles/owner",
      "members": [
        "user:mike@grchive.com"
      ]
    }
  ]
}
		`)
			body := ioutil.NopCloser(strings.NewReader(data))
			return &http.Response{
				StatusCode: http.StatusOK,
				Body:       body,
			}, nil

		},
		RolePermissions: map[string]gcloud_utility.MockGCloudFn{
			"roles/owner": func() (*http.Response, error) {
				data := `
{
  "name": "roles/owner",
  "title": "Owner",
  "description": "Created on: 2020-08-19",
  "includedPermissions": [
    "iam.roles.get",
    "resourcemanager.projects.getIamPolicy"
  ],
  "etag": "BwWtPQ96Oqs="
}
				`
				body := ioutil.NopCloser(strings.NewReader(data))
				return &http.Response{
					StatusCode: http.StatusOK,
					Body:       body,
				}, nil
			},
			"organizations/94248544035/roles/CustomRole294": func() (*http.Response, error) {
				data := `
{
  "name": "organizations/94248544035/roles/CustomRole294",
  "title": "Custom Role 294",
  "description": "Created on: 2020-08-19",
  "includedPermissions": [
    "storage.objects.create",
    "storage.objects.delete",
    "storage.objects.get",
    "storage.objects.list"
  ],
  "etag": "BwWtPQ96Oqs="
}
				`
				body := ioutil.NopCloser(strings.NewReader(data))
				return &http.Response{
					StatusCode: http.StatusOK,
					Body:       body,
				}, nil

			},
		},
	}
	conn, err := CreateGCloudConnector(&EtlGCloudOptions{
		Client:    client,
		ProjectId: "test",
	})
	g.Expect(err).To(gomega.BeNil())

	itf, err := conn.GetUserInterface()
	g.Expect(err).To(gomega.BeNil())

	users, source, err := itf.GetUserListing()
	g.Expect(err).To(gomega.BeNil())

	g.Expect(source).NotTo(gomega.BeNil())
	g.Expect(len(source.Commands)).To(gomega.Equal(3))

	refUsers := map[string]*types.EtlUser{
		"user:mike@grchive.com": &types.EtlUser{
			Username: "user:mike@grchive.com",
			Email:    "mike@grchive.com",
			Roles: map[string]*types.EtlRole{
				"roles/owner": &types.EtlRole{
					Name: "Owner",
					Permissions: map[string][]string{
						"Self": []string{
							"iam.roles.get",
							"resourcemanager.projects.getIamPolicy",
						},
					},
				},
				"organizations/94248544035/roles/CustomRole294": &types.EtlRole{
					Name: "Custom Role 294",
					Permissions: map[string][]string{
						"Self": []string{
							"storage.objects.create",
							"storage.objects.delete",
							"storage.objects.get",
							"storage.objects.list",
						},
					},
				},
			},
		},

		"serviceAccount:grchive-service-account@grchive-v3.iam.gserviceaccount.com": &types.EtlUser{
			Username: "serviceAccount:grchive-service-account@grchive-v3.iam.gserviceaccount.com",
			Email:    "grchive-service-account@grchive-v3.iam.gserviceaccount.com",
			Roles: map[string]*types.EtlRole{
				"organizations/94248544035/roles/CustomRole294": &types.EtlRole{
					Name: "Custom Role 294",
					Permissions: map[string][]string{
						"Self": []string{
							"storage.objects.create",
							"storage.objects.delete",
							"storage.objects.get",
							"storage.objects.list",
						},
					},
				},
			},
		},
	}

	test_utility.CompareUserListing(g, users, refUsers, test_utility.CompareUserListingOptions{})
}
