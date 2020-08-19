package ldap

import (
	"fmt"
	"github.com/go-ldap/ldap/v3"
	"gitlab.com/grchive/grchive-v3/shared/etl/connectors"
	"gitlab.com/grchive/grchive-v3/shared/etl/types"
	"strings"
)

const ldapAttributeKeyConstantPrefix = "@CONSTANT@"

type EtlLdapConnectorUser struct {
	opts *EtlLdapOptions
}

func createLdapConnectorUser(opts *EtlLdapOptions) (*EtlLdapConnectorUser, error) {
	return &EtlLdapConnectorUser{
		opts: opts,
	}, nil
}

func parseAttributeJoin(keys []string, attrs map[string][]string) string {
	ret := make([]string, len(keys))
	for idx, k := range keys {
		if strings.HasPrefix(k, ldapAttributeKeyConstantPrefix) {
			ret[idx] = strings.TrimPrefix(k, ldapAttributeKeyConstantPrefix)
		} else {
			values, ok := attrs[k]
			if !ok {
				continue
			}
			ret[idx] = strings.TrimSpace(strings.Join(values, ", "))
		}
	}
	return strings.TrimSpace(strings.Join(ret, ""))
}

func createRawDataFromLdapEntry(e *ldap.Entry) string {
	raw := strings.Builder{}
	raw.WriteString(fmt.Sprintf("dn: %s\n", e.DN))
	for _, attr := range e.Attributes {
		for _, v := range attr.Values {
			raw.WriteString(fmt.Sprintf("%s: %s\n", attr.Name, v))
		}
	}
	return raw.String()
}

func createEtlUserFromLdapEntry(e *ldap.Entry, cfg EtlLdapUserConfig) *types.EtlUser {
	user := types.EtlUser{}

	attributeMap := map[string][]string{}
	for _, attr := range e.Attributes {
		attributeMap[attr.Name] = attr.Values
	}

	user.Username = parseAttributeJoin(cfg.UsernameAttribute, attributeMap)
	user.FullName = parseAttributeJoin(cfg.FullNameAttributes, attributeMap)
	user.Email = parseAttributeJoin(cfg.EmailAttributes, attributeMap)
	return &user
}

func (c *EtlLdapConnectorUser) GetUserListing() ([]*types.EtlUser, *connectors.EtlSourceInfo, error) {
	searchReq := ldap.SearchRequest{
		BaseDN:       fmt.Sprintf("%s,%s", c.opts.Config.User.ParentDn, c.opts.Config.RootDn),
		Scope:        ldap.ScopeSingleLevel,
		DerefAliases: ldap.NeverDerefAliases,
		SizeLimit:    0,
		TimeLimit:    0,
		TypesOnly:    false,
		Filter:       "(objectclass=*)",
	}

	result, err := c.opts.Client.Search(&searchReq)
	if err != nil {
		return nil, nil, err
	}

	retUsers := []*types.EtlUser{}
	rawData := strings.Builder{}
	for _, entry := range result.Entries {
		rawData.WriteString(createRawDataFromLdapEntry(entry))
		rawData.WriteString("\n")
		retUsers = append(retUsers, createEtlUserFromLdapEntry(entry, c.opts.Config.User))
	}

	source := connectors.CreateSourceInfo()
	// Rebuild the command to be as class to the expected ldapsearch equivalent command.
	cmd := connectors.EtlCommandInfo{
		Command: fmt.Sprintf("ldapsearch -s one -b \"%s\" -a never -l 0 -z -0 '(objectclass=*)'", searchReq.BaseDN),
		RawData: rawData.String(),
	}
	source.AddCommand(&cmd)

	return retUsers, source, nil
}
