package ldap

import (
	"github.com/go-ldap/ldap/v3"
	"github.com/onsi/gomega"
	"testing"
)

func TestCreateLdapConnector(t *testing.T) {
	g := gomega.NewGomegaWithT(t)

	client := &ldap.Conn{}
	refDn := "dndndndn"
	refUserConfig := EtlLdapUserConfig{
		ParentDn:           "asdf",
		UsernameAttribute:  []string{"1", "2"},
		FullNameAttributes: []string{"3", "4"},
		EmailAttributes:    []string{"5", "6", "7"},
	}

	refOptions := EtlLdapOptions{
		Client: client,
		Config: EtlLdapConfig{
			RootDn: refDn,
			User:   refUserConfig,
		},
	}

	conn, err := CreateLdapConnector(&refOptions)
	g.Expect(err).To(gomega.BeNil())
	g.Expect(conn).NotTo(gomega.BeNil())
	g.Expect(conn.opts).NotTo(gomega.BeNil())
	g.Expect(*conn.opts).To(gomega.Equal(refOptions))

	g.Expect(conn.users).NotTo(gomega.BeNil())
	g.Expect(conn.users.opts).To(gomega.Equal(conn.opts))
	g.Expect(conn.GetUserInterface()).To(gomega.Equal(conn.users))
}
