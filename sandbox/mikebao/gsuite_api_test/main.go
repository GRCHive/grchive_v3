package main

import (
	"flag"
	"fmt"
	"gitlab.com/grchive/grchive-v3/shared/etl/connectors/saas/gsuite"
	"gitlab.com/grchive/grchive-v3/shared/utility/http"
	"golang.org/x/net/context"
	"golang.org/x/oauth2/google"
	"google.golang.org/api/admin/directory/v1"
	"io/ioutil"
)

var credentialFname string

func main() {
	flag.StringVar(&credentialFname, "cred", "", "Google OAuth Client credentials")
	flag.Parse()

	credentials, err := ioutil.ReadFile(credentialFname)
	if err != nil {
		fmt.Printf("Read Credential Error: %s\n", err.Error())
		return
	}

	config, err := google.JWTConfigFromJSON(credentials, admin.AdminDirectoryUserReadonlyScope)
	if err != nil {
		fmt.Printf("Generate Config Error: %s\n", err.Error())
		return
	}
	config.Subject = "mike@grchive.com"
	ts := config.TokenSource(context.Background())

	connector, err := gsuite.CreateGSuiteConnector(&gsuite.EtlGSuiteOptions{
		Client:     http_utility.CreateOAuth2AuthorizedClient(ts),
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
