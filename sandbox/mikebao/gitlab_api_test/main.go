package main

import (
	"flag"
	"fmt"
	"gitlab.com/grchive/grchive-v3/shared/etl/connectors/saas/gitlab"
	"gitlab.com/grchive/grchive-v3/shared/utility/auth"
)

var clientId string
var clientSecret string

func main() {
	flag.StringVar(&clientId, "id", "", "Gitlab OAuth Client Id")
	flag.StringVar(&clientSecret, "secret", "", "Gitlab OAuth Client Secret")
	flag.Parse()

	config := auth_utility.CreateGitlabOAuthConfig(clientId, clientSecret, "http://localhost", "api")
	fmt.Printf("GO HERE: %s\n", config.AuthCodeURL(""))

	fmt.Printf("COPY AND PASTE AUTH CODE: \n")
	code := ""
	_, err := fmt.Scan(&code)
	if err != nil {
		fmt.Printf("Read Code Error: %s\n", err.Error())
		return
	}

	fmt.Printf("CODE: %s\n", code)

	ts, err := auth_utility.CreateGitlabOAuthTokenSource(config, code)
	if err != nil {
		fmt.Printf("Create Token Source Error: %s\n", err.Error())
		return
	}

	connector, err := gitlab.CreateGitlabConnector(&gitlab.EtlGitlabOptions{
		Client:  auth_utility.CreateGitlabHttpClient(ts),
		GroupId: "grchive",
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
