package main

import (
	"flag"
	"fmt"
	"gitlab.com/grchive/grchive-v3/shared/etl/connectors/saas/bitbucket"
	"gitlab.com/grchive/grchive-v3/shared/utility/auth"
)

var clientId string
var clientSecret string

func main() {
	flag.StringVar(&clientId, "id", "", "Bitbucket OAuth Client Id")
	flag.StringVar(&clientSecret, "secret", "", "Bitbucket OAuth Client Secret")
	flag.Parse()

	config := auth_utility.CreateBitbucketOAuthConfig(clientId, clientSecret, "http://localhost", "account")
	fmt.Printf("GO HERE: %s\n", config.AuthCodeURL(""))

	fmt.Printf("COPY AND PASTE AUTH CODE: \n")
	code := ""
	_, err := fmt.Scan(&code)
	if err != nil {
		fmt.Printf("Read Code Error: %s\n", err.Error())
		return
	}

	fmt.Printf("CODE: %s\n", code)

	ts, err := auth_utility.CreateBitbucketOAuthTokenSource(config, code)
	if err != nil {
		fmt.Printf("Create Token Source Error: %s\n", err.Error())
		return
	}

	connector, err := bitbucket.CreateBitbucketConnector(&bitbucket.EtlBitbucketOptions{
		Client:      auth_utility.CreateBitbucketHttpClient(ts),
		WorkspaceId: "grchive",
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
