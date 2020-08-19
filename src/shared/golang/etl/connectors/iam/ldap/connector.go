package ldap

import (
	"github.com/go-ldap/ldap/v3"
	"gitlab.com/grchive/grchive-v3/shared/etl/connectors"
)

type EtlLdapUserConfig struct {
	ParentDn           string
	UsernameAttribute  []string
	FullNameAttributes []string
	EmailAttributes    []string
}

type EtlLdapConfig struct {
	RootDn string

	// Data parsing options
	User EtlLdapUserConfig
}

type EtlLdapOptions struct {
	Client ldap.Client
	Config EtlLdapConfig
}

type EtlLdapConnector struct {
	opts  *EtlLdapOptions
	users *EtlLdapConnectorUser
}

func (c *EtlLdapConnector) GetUserInterface() (connectors.EtlConnectorUserInterface, error) {
	return c.users, nil
}

func CreateLdapConnector(opts *EtlLdapOptions) (*EtlLdapConnector, error) {
	var err error
	ret := EtlLdapConnector{
		opts: opts,
	}
	ret.users, err = createLdapConnectorUser(opts)

	if err != nil {
		return nil, err
	}

	return &ret, nil
}
