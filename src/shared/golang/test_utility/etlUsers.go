package test_utility

import (
	"github.com/onsi/gomega"
	"gitlab.com/grchive/grchive-v3/shared/etl/types"
	"sort"
	"strings"
	"time"
)

type CompareUserListingOptions struct {
	UsersToIgnore             []string
	PermissionObjectsToIgnore []string
}

func CompareUserListing(g *gomega.GomegaWithT, users []*types.EtlUser, refUsers map[string]*types.EtlUser, opts CompareUserListingOptions) {
	g.Expect(len(users)).To(gomega.Equal(len(refUsers) + len(opts.UsersToIgnore)))
USER:
	for _, u := range users {
		for _, iu := range opts.UsersToIgnore {
			if u.Username == iu {
				continue USER
			}
		}

		refU, ok := refUsers[u.Username]
		g.Expect(ok).To(gomega.BeTrue(), "Finding username: "+u.Username)

		g.Expect(u.Username).To(gomega.Equal(refU.Username))
		g.Expect(u.FullName).To(gomega.Equal(refU.FullName))
		g.Expect(u.Email).To(gomega.Equal(refU.Email))

		if refU.CreatedTime == nil {
			g.Expect(u.CreatedTime).To(gomega.BeNil())
		} else {
			g.Expect(u.CreatedTime).NotTo(gomega.BeNil())
			g.Expect(*u.CreatedTime).To(gomega.BeTemporally("~", *refU.CreatedTime, time.Second))
		}

		if refU.LastChangeTime == nil {
			g.Expect(u.LastChangeTime).To(gomega.BeNil())
		} else {
			g.Expect(u.LastChangeTime).NotTo(gomega.BeNil())
			g.Expect(*u.LastChangeTime).To(gomega.BeTemporally("~", *refU.LastChangeTime, time.Second))
		}

		g.Expect(len(u.Roles)).To(gomega.Equal(len(refU.Roles)))

		for rkey, r := range u.Roles {
			refRole, ok := refU.Roles[rkey]
			g.Expect(ok).To(gomega.BeTrue(), "Finding role: "+rkey)
			g.Expect(r.Name).To(gomega.Equal(refRole.Name))
			g.Expect(len(r.Permissions)).To(gomega.Equal(len(refRole.Permissions)))

		PERM:
			for object, permissions := range r.Permissions {
				for _, o := range opts.PermissionObjectsToIgnore {
					if strings.Contains(object, o) {
						continue PERM
					}
				}

				refPermissions, ok := refRole.Permissions[object]
				g.Expect(ok).To(gomega.BeTrue(), "Failed to find ref permissions: "+object)

				sort.Strings(permissions)
				sort.Strings(refPermissions)
				g.Expect(len(permissions)).To(gomega.Equal(len(refPermissions)))
				for i, p := range permissions {
					g.Expect(p).To(gomega.Equal(refPermissions[i]))
				}
			}
		}
	}
}
