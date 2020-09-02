package main

import (
	"flag"
	"fmt"
	"gitlab.com/grchive/grchive-v3/shared/etl/connectors/saas/office365"
	"gitlab.com/grchive/grchive-v3/shared/utility/auth"
	"golang.org/x/oauth2"
)

var tenant string
var clientId string
var clientSecret string

func setupTokenSource(resource string, scopes ...string) (oauth2.TokenSource, error) {
	config := auth_utility.AzureOAuthSetup{
		Tenant:       tenant,
		ClientId:     clientId,
		ClientSecret: clientSecret,
		RedirectURL:  "http://localhost",
		Scopes:       scopes,
	}.ToOAuthConfig()

	fmt.Printf("GO HERE: %s\n", config.AuthCodeURL(""))

	fmt.Printf("COPY AND PASTE AUTH CODE: \n")
	code := ""
	_, err := fmt.Scan(&code)
	if err != nil {
		return nil, err
	}

	fmt.Printf("CODE: %s\n", code)

	tokenSources, err := auth_utility.CreateAzureOAuthTokenSource(config, code)
	if err != nil {
		return nil, err
	}

	return tokenSources[resource], nil
}

func main() {
	flag.StringVar(&tenant, "tenant", "", "Azure Tenant")
	flag.StringVar(&clientId, "id", "", "Azure App OAuth Client Id")
	flag.StringVar(&clientSecret, "secret", "", "Azure App OAuth Client Secret")
	flag.Parse()

	graphTs, err := setupTokenSource(auth_utility.AzureGraphResource,
		"offline_access",
		"https://graph.microsoft.com/user.read.all",
		"https://graph.microsoft.com/rolemanagement.read.directory",
	)
	if err != nil {
		fmt.Printf("Graph Token Error: %s\n", err.Error())
		return
	}

	connector, err := office365.CreateOffice365Connector(&office365.EtlOffice365Options{
		Client: auth_utility.CreateAzureHttpClient(graphTs),
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
		for _, r := range u.Roles {
			fmt.Printf("\t%+v\n", r)
		}
	}

	fmt.Printf("=================== SOURCES ===================\n")
	for _, s := range etlSource.Commands {
		fmt.Printf("%+v\n", s)
	}
}
