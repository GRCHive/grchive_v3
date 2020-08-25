package main

import (
	"flag"
	"fmt"
	"gitlab.com/grchive/grchive-v3/shared/etl/connectors/paas/heroku"
	"gitlab.com/grchive/grchive-v3/shared/utility/auth"
)

var clientId string
var clientSecret string
var team string

func main() {
	flag.StringVar(&clientId, "id", "", "Heroku OAuth Client Id")
	flag.StringVar(&clientSecret, "secret", "", "Heroku OAuth Client Secret")
	flag.StringVar(&team, "team", "", "Heroku Team")
	flag.Parse()

	// Need read_write to access grants for some reason.
	config := auth_utility.CreateHerokuOAuthConfig(clientId, clientSecret, "http://localhost", "global")
	fmt.Printf("GO HERE: %s\n", config.AuthCodeURL(""))

	fmt.Printf("COPY AND PASTE AUTH CODE: \n")
	code := ""
	_, err := fmt.Scan(&code)
	if err != nil {
		fmt.Printf("Read Code Error: %s\n", err.Error())
		return
	}

	fmt.Printf("CODE: %s\n", code)

	ts, err := auth_utility.CreateHerokuOAuthTokenSource(config, code)
	if err != nil {
		fmt.Printf("Create Token Source Error: %s\n", err.Error())
		return
	}

	connector, err := heroku.CreateHerokuConnector(&heroku.EtlHerokuOptions{
		Client:   auth_utility.CreateHerokuHttpClient(ts),
		TeamName: team,
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
