package main

import (
	"flag"
	"fmt"
	"gitlab.com/grchive/grchive-v3/shared/etl/connectors/iam/auth0"
	"gitlab.com/grchive/grchive-v3/shared/utility/auth"
)

var domain string
var audience string
var clientId string
var clientSecret string

func main() {
	flag.StringVar(&domain, "domain", "", "Auth0 Domain")
	flag.StringVar(&audience, "audience", "", "Auth0 API Audience")
	flag.StringVar(&clientId, "id", "", "Auth0 Client ID")
	flag.StringVar(&clientSecret, "secret", "", "Auth0 Client Secret")
	flag.Parse()

	var err error
	client := &auth_utility.OAuthClient{}
	if clientId == "" || clientSecret == "" {
		client, err = auth_utility.Auth0RegisterClient(domain, auth_utility.OAuthClientRegistration{
			ClientName:   "Auth0 Sandbox",
			RedirectUris: []string{"http://localhost"},
		})
		if err != nil {
			fmt.Printf("Register client error: %s\n", err.Error())
			return
		}
	} else {
		client.ClientId = clientId
		client.ClientSecret = clientSecret
		client.RedirectUris = []string{"http://localhost"}
	}

	fmt.Printf("CLIENT: %+v\n", *client)

	ts, err := auth_utility.CreateAuth0OAuthTokenSource(domain, audience, client)
	if err != nil {
		fmt.Printf("Create Token Source Error: %s\n", err.Error())
		return
	}

	connector, err := auth0.CreateAuth0Connector(&auth0.EtlAuth0Options{
		Client: auth_utility.CreateAuth0HttpClient(ts),
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
