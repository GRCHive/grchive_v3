package auth_utility

import (
	"crypto/tls"
	"github.com/go-ldap/ldap/v3"
)

func CreateLDAPClient(addr string, tlsConfig *tls.Config) (ldap.Client, error) {
	opts := []ldap.DialOpt{}

	if tlsConfig != nil {
		opts = append(opts, ldap.DialWithTLSConfig(tlsConfig))
	}

	return ldap.DialURL(addr, opts...)
}
