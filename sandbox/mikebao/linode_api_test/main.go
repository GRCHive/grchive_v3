package main

import (
	"flag"
	"fmt"
	"gitlab.com/grchive/grchive-v3/shared/etl/connectors/iaas/linode"
	"gitlab.com/grchive/grchive-v3/shared/utility/auth"
)

var clientId string
var clientSecret string

func main() {
	flag.StringVar(&clientId, "id", "", "Gitlab OAuth Client Id")
	flag.StringVar(&clientSecret, "secret", "", "Gitlab OAuth Client Secret")
	flag.Parse()

	// Need read_write to access grants for some reason.
	config := auth_utility.CreateLinodeOAuthConfig(clientId, clientSecret, "http://localhost", "account:read_write")
	fmt.Printf("GO HERE: %s\n", config.AuthCodeURL(""))

	fmt.Printf("COPY AND PASTE AUTH CODE: \n")
	code := ""
	_, err := fmt.Scan(&code)
	if err != nil {
		fmt.Printf("Read Code Error: %s\n", err.Error())
		return
	}

	fmt.Printf("CODE: %s\n", code)

	ts, err := auth_utility.CreateLinodeOAuthTokenSource(config, code)
	if err != nil {
		fmt.Printf("Create Token Source Error: %s\n", err.Error())
		return
	}

	connector, err := linode.CreateLinodeConnector(&linode.EtlLinodeOptions{
		Client: auth_utility.CreateLinodeHttpClient(ts),
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
