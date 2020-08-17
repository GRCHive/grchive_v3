package main

import (
	"flag"
	"fmt"
	"gitlab.com/grchive/grchive-v3/shared/etl/connectors/saas/gsuite"
	"gitlab.com/grchive/grchive-v3/shared/utility/auth"
	"google.golang.org/api/admin/directory/v1"
)

var credentialFname string

func main() {
	flag.StringVar(&credentialFname, "cred", "", "Google OAuth Client credentials")
	flag.Parse()

	ts, err := auth_utility.CreateGSuiteOAuthTokenSource(credentialFname, "mike@grchive.com", admin.AdminDirectoryUserReadonlyScope)
	if err != nil {
		fmt.Printf("Create Token Source Error: %s\n", err.Error())
		return
	}

	connector, err := gsuite.CreateGSuiteConnector(&gsuite.EtlGSuiteOptions{
		Client:     auth_utility.CreateGSuiteHttpClient(ts),
		CustomerId: "01v99vjw",
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
