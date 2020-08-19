package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"gitlab.com/grchive/grchive-v3/shared/etl/connectors/iam/ldap"
	"gitlab.com/grchive/grchive-v3/shared/utility/auth"
	"io/ioutil"
)

var configFname string

type Config struct {
	Uri    string
	Config ldap.EtlLdapConfig
}

func readConfigFromFile(fname string) (*Config, error) {
	data, err := ioutil.ReadFile(fname)
	if err != nil {
		return nil, err
	}

	config := Config{}
	err = json.Unmarshal(data, &config)
	if err != nil {
		return nil, err
	}

	return &config, nil
}

func main() {
	flag.StringVar(&configFname, "config", "", "LDAP JSON Config Filename")
	flag.Parse()

	config, err := readConfigFromFile(configFname)
	if err != nil {
		fmt.Printf("Read Config Error: %s\n", err.Error())
		return
	}

	client, err := auth_utility.CreateLDAPClient(config.Uri, nil)
	if err != nil {
		fmt.Printf("Connect LDAP Error: %s\n", err.Error())
		return
	}

	connector, err := ldap.CreateLdapConnector(&ldap.EtlLdapOptions{
		Client: client,
		Config: config.Config,
	})
	if err != nil {
		fmt.Printf("Create Connector Error: %s\n", err.Error())
		return
	}

	users, err := connector.GetUserInterface()
	if err != nil {
		fmt.Printf("Get User Interface Error: %s\n", err.Error())
		return
	}

	etlUsers, etlSource, err := users.GetUserListing()
	if err != nil {
		fmt.Printf("Get User Listing Error: %s\n", err.Error())
		return
	}

	fmt.Printf("=================== USERS ===================\n")
	for _, u := range etlUsers {
		fmt.Printf("%+v\n", u)
	}

	fmt.Printf("=================== SOURCES ===================\n")
	for _, s := range etlSource.Commands {
		fmt.Printf("%+v\n", s)
	}
}
