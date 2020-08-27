package main

import (
	"flag"
	"fmt"
	"gitlab.com/grchive/grchive-v3/shared/etl/connectors/iam/okta"
	"gitlab.com/grchive/grchive-v3/shared/utility/auth"
)

var token string
var domain string

func main() {
	flag.StringVar(&token, "token", "", "Okta Auth Token")
	flag.StringVar(&domain, "domain", "", "Okta Domain")
	flag.Parse()

	connector, err := okta.CreateOktaConnector(&okta.EtlOktaOptions{
		Client: auth_utility.CreateOktaHttpClient(token),
		Domain: domain,
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
