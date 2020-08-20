package main

import (
	"flag"
	"fmt"
	"gitlab.com/grchive/grchive-v3/shared/etl/connectors/iaas/aws"
	"gitlab.com/grchive/grchive-v3/shared/utility/auth"
	"gitlab.com/grchive/grchive-v3/shared/utility/time"
)

var keyId string
var keySecret string

func main() {
	flag.StringVar(&keyId, "id", "", "AWS Access Key ID")
	flag.StringVar(&keySecret, "secret", "", "AWS Access Key Secret")
	flag.Parse()

	connector, err := aws.CreateAWSConnector(&aws.EtlAWSOptions{
		Client: auth_utility.CreateAWSHttpClient(time_utility.RealClock{}, keyId, keySecret),
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
