package main

import (
	"flag"
	"fmt"
	"gitlab.com/grchive/grchive-v3/shared/etl/connectors/saas/gcloud"
	"gitlab.com/grchive/grchive-v3/shared/utility/auth"
	"google.golang.org/api/cloudresourcemanager/v1"
	"google.golang.org/api/iam/v1"
)

var credentialFname string

func main() {
	flag.StringVar(&credentialFname, "cred", "", "Google OAuth Client credentials")
	flag.Parse()

	ts, err := auth_utility.CreateGoogleOAuthTokenSource(
		credentialFname,
		"",
		cloudresourcemanager.CloudPlatformReadOnlyScope,
		iam.CloudPlatformScope,
	)
	if err != nil {
		fmt.Printf("Create Token Source Error: %s\n", err.Error())
		return
	}

	connector, err := gcloud.CreateGCloudConnector(&gcloud.EtlGCloudOptions{
		Client:    auth_utility.CreateGoogleHttpClient(ts),
		ProjectId: "grchive-v3",
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
