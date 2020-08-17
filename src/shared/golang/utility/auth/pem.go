package auth_utility

import (
	"crypto/rsa"
	"crypto/x509"
	"encoding/pem"
	"errors"
	"io/ioutil"
)

func ReadPEMFile(fname string) (*pem.Block, error) {
	raw, err := ioutil.ReadFile(fname)
	if err != nil {
		return nil, err
	}

	pemData, _ := pem.Decode(raw)
	return pemData, nil
}

func ReadRSAPrivateKeyFromPEM(fname string) (*rsa.PrivateKey, error) {
	pemBlock, err := ReadPEMFile(fname)
	if err != nil {
		return nil, err
	}

	if pemBlock.Type != "RSA PRIVATE KEY" {
		return nil, errors.New("Input key does not have type of RSA PRIVATE KEY.")
	}

	key, err := x509.ParsePKCS1PrivateKey(pemBlock.Bytes)
	if err != nil {
		return nil, err
	}

	return key, nil
}
