package ldap

import (
	"github.com/go-ldap/ldap/v3"
	"github.com/onsi/gomega"
	"gitlab.com/grchive/grchive-v3/shared/etl/types"
	"gitlab.com/grchive/grchive-v3/shared/test_utility"
	"gitlab.com/grchive/grchive-v3/tests/shared/etl/connectors/iam/ldap_utility"
	"testing"
)

func TestParseAttributeJoin(t *testing.T) {
	g := gomega.NewGomegaWithT(t)

	attrs := map[string][]string{
		"test1": []string{"test1"},
		"test2": []string{},
		"test3": []string{"test2", "test3"},
		"test4": []string{"test4"},
	}

	for _, test := range []struct {
		Keys []string
		Ref  string
	}{
		{
			Keys: []string{"test1"},
			Ref:  "test1",
		},
		{
			Keys: []string{"test2"},
			Ref:  "",
		},
		{
			Keys: []string{"test3"},
			Ref:  "test2, test3",
		},
		{
			Keys: []string{"test1", "test4"},
			Ref:  "test1test4",
		},
		{
			Keys: []string{"test1", "test2"},
			Ref:  "test1",
		},
		{
			Keys: []string{"test1", "test3"},
			Ref:  "test1test2, test3",
		},
		{
			Keys: []string{"@CONSTANT@blah"},
			Ref:  "blah",
		},
	} {
		cmp := parseAttributeJoin(test.Keys, attrs)
		g.Expect(cmp).To(gomega.Equal(test.Ref))
	}
}

func TestCreateRawDataFromLdapEntry(t *testing.T) {
	g := gomega.NewGomegaWithT(t)

	for _, test := range []struct {
		Entry *ldap.Entry
		Ref   string
	}{
		{
			Entry: &ldap.Entry{
				DN: "cn=test,dc=grchive,dc=com",
				Attributes: []*ldap.EntryAttribute{
					ldap.NewEntryAttribute("test1", []string{"Michael"}),
					ldap.NewEntryAttribute("test2", []string{"Bao"}),
					ldap.NewEntryAttribute("test3", []string{"mike@grchive.com"}),
					ldap.NewEntryAttribute("test4", []string{"mike"}),
					ldap.NewEntryAttribute("test5", []string{"null", "two"}),
				},
			},
			Ref: `dn: cn=test,dc=grchive,dc=com
test1: Michael
test2: Bao
test3: mike@grchive.com
test4: mike
test5: null
test5: two
`,
		},
		{
			Entry: &ldap.Entry{
				DN:         "test",
				Attributes: []*ldap.EntryAttribute{},
			},
			Ref: `dn: test
`,
		},
	} {
		cmp := createRawDataFromLdapEntry(test.Entry)
		g.Expect(cmp).To(gomega.Equal(test.Ref))
	}
}

func TestCreateEtlUserFromLdapEntry(t *testing.T) {
	g := gomega.NewGomegaWithT(t)

	for _, test := range []struct {
		Entry *ldap.Entry
		Cfg   EtlLdapUserConfig
		Ref   types.EtlUser
	}{
		{
			Entry: &ldap.Entry{
				DN: "cn=test,dc=grchive,dc=com",
				Attributes: []*ldap.EntryAttribute{
					ldap.NewEntryAttribute("test1", []string{"Michael"}),
					ldap.NewEntryAttribute("test2", []string{"Bao"}),
					ldap.NewEntryAttribute("test3", []string{"mike@grchive.com"}),
					ldap.NewEntryAttribute("test4", []string{"mike"}),
					ldap.NewEntryAttribute("test5", []string{"null", "two"}),
				},
			},
			Cfg: EtlLdapUserConfig{
				UsernameAttribute:  []string{"test4"},
				FullNameAttributes: []string{"test1", "@CONSTANT@ ", "test2"},
				EmailAttributes:    []string{"test3"},
			},
			Ref: types.EtlUser{
				Username: "mike",
				FullName: "Michael Bao",
				Email:    "mike@grchive.com",
			},
		},

		{
			Entry: &ldap.Entry{
				DN: "cn=test,dc=grchive,dc=com",
				Attributes: []*ldap.EntryAttribute{
					ldap.NewEntryAttribute("test1", []string{"Michael"}),
					ldap.NewEntryAttribute("test2", []string{"Bao"}),
					ldap.NewEntryAttribute("test3", []string{"mike@grchive.com"}),
					ldap.NewEntryAttribute("test4", []string{"mike"}),
					ldap.NewEntryAttribute("test5", []string{"null", "two"}),
				},
			},
			Cfg: EtlLdapUserConfig{
				UsernameAttribute:  []string{},
				FullNameAttributes: []string{"test5"},
				EmailAttributes:    []string{"test4", "@CONSTANT@@gmail.com"},
			},
			Ref: types.EtlUser{
				Username: "",
				FullName: "null, two",
				Email:    "mike@gmail.com",
			},
		},
	} {
		cmp := createEtlUserFromLdapEntry(test.Entry, test.Cfg)
		g.Expect(*cmp).To(gomega.Equal(test.Ref))
	}
}

func TestUserListingParse(t *testing.T) {
	g := gomega.NewGomegaWithT(t)

	client := &ldap_utility.MockLdapClient{
		UserData: []*ldap.Entry{
			&ldap.Entry{
				DN: "1",
				Attributes: []*ldap.EntryAttribute{
					ldap.NewEntryAttribute("test1", []string{"Michael"}),
					ldap.NewEntryAttribute("test2", []string{"Bao"}),
					ldap.NewEntryAttribute("test3", []string{"mike@grchive.com"}),
					ldap.NewEntryAttribute("test4", []string{"mike"}),
					ldap.NewEntryAttribute("test5", []string{"null", "two"}),
				},
			},
			&ldap.Entry{
				DN: "2",
				Attributes: []*ldap.EntryAttribute{
					ldap.NewEntryAttribute("test1", []string{"Derek"}),
					ldap.NewEntryAttribute("test2", []string{"Chin"}),
					ldap.NewEntryAttribute("test4", []string{"derek"}),
					ldap.NewEntryAttribute("test5", []string{}),
				},
			},
		},
	}

	for _, config := range []EtlLdapConfig{
		EtlLdapConfig{
			User: EtlLdapUserConfig{
				UsernameAttribute:  []string{"test4"},
				FullNameAttributes: []string{"test1", "test2"},
				EmailAttributes:    []string{"test3"},
			},
		},
		EtlLdapConfig{
			User: EtlLdapUserConfig{
				UsernameAttribute:  []string{"test5"},
				FullNameAttributes: []string{},
				EmailAttributes:    []string{"test4", "@CONSTANT@@grchive.com"},
			},
		},
	} {
		refUsers := map[string]*types.EtlUser{}
		for _, d := range client.UserData {
			refU := createEtlUserFromLdapEntry(d, config.User)
			refUsers[refU.Username] = refU
		}

		conn, err := CreateLdapConnector(&EtlLdapOptions{
			Client: client,
			Config: config,
		})
		g.Expect(err).To(gomega.BeNil())

		itf, err := conn.GetUserInterface()
		g.Expect(err).To(gomega.BeNil())

		users, source, err := itf.GetUserListing()
		g.Expect(err).To(gomega.BeNil())

		g.Expect(source).NotTo(gomega.BeNil())
		g.Expect(len(source.Commands)).To(gomega.Equal(1))

		test_utility.CompareUserListing(g, users, refUsers, test_utility.CompareUserListingOptions{})
	}

}
