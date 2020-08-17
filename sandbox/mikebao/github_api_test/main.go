package main

import (
	"flag"
	"fmt"
	"gitlab.com/grchive/grchive-v3/shared/etl/connectors/saas/github"
	"gitlab.com/grchive/grchive-v3/shared/utility/auth"
	"gitlab.com/grchive/grchive-v3/shared/utility/time"
)

var privateKeyFname string
var appId string
var installationId string

func main() {
	flag.StringVar(&privateKeyFname, "key", "", "Github App Private Key Filename")
	flag.StringVar(&appId, "app", "", "Github App ID")
	flag.StringVar(&installationId, "install", "", "Github Installation ID")
	flag.Parse()

	clock := time_utility.RealClock{}
	token, err := auth_utility.CreateGithubJWTToken(clock, appId, privateKeyFname)
	if err != nil {
		fmt.Printf("Create Token Source Error: %s\n", err.Error())
		return
	}

	fmt.Printf("TOKEN: %s\n", token)
	setupConnector, err := github.CreateGithubConnector(&github.EtlGithubOptions{
		Client: auth_utility.CreateGithubHttpJWTClient(token),
		OrgId:  "GRCHive",
	})

	accessToken, err := setupConnector.GetInstallationAccessToken(installationId)
	if err != nil {
		fmt.Printf("Get Access Token: %s\n", err.Error())
		return
	}

	fmt.Printf("ACCESS: %s\n", accessToken)
	connector, err := github.CreateGithubConnector(&github.EtlGithubOptions{
		Client: auth_utility.CreateGithubHttpInstallationClient(accessToken),
		OrgId:  "GRCHive",
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
