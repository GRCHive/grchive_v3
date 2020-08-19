package ldap_utility

import (
	"crypto/tls"
	"github.com/go-ldap/ldap/v3"
	"time"
)

type MockLdapClient struct {
	UserData []*ldap.Entry
}

func (c *MockLdapClient) Start()                           {}
func (c *MockLdapClient) StartTLS(*tls.Config) error       { return nil }
func (c *MockLdapClient) Close()                           {}
func (c *MockLdapClient) SetTimeout(time.Duration)         {}
func (c *MockLdapClient) Bind(string, string) error        { return nil }
func (c *MockLdapClient) UnauthenticatedBind(string) error { return nil }
func (c *MockLdapClient) SimpleBind(*ldap.SimpleBindRequest) (*ldap.SimpleBindResult, error) {
	return nil, nil
}
func (c *MockLdapClient) ExternalBind() error                          { return nil }
func (c *MockLdapClient) Add(*ldap.AddRequest) error                   { return nil }
func (c *MockLdapClient) Del(*ldap.DelRequest) error                   { return nil }
func (c *MockLdapClient) Modify(*ldap.ModifyRequest) error             { return nil }
func (c *MockLdapClient) ModifyDN(*ldap.ModifyDNRequest) error         { return nil }
func (c *MockLdapClient) Compare(string, string, string) (bool, error) { return false, nil }
func (c *MockLdapClient) PasswordModify(*ldap.PasswordModifyRequest) (*ldap.PasswordModifyResult, error) {
	return nil, nil
}
func (c *MockLdapClient) Search(*ldap.SearchRequest) (*ldap.SearchResult, error) {
	return &ldap.SearchResult{
		Entries: c.UserData,
	}, nil
}
func (c *MockLdapClient) SearchWithPaging(*ldap.SearchRequest, uint32) (*ldap.SearchResult, error) {
	return nil, nil
}
