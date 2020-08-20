package aws

import (
	"errors"
	"fmt"
	"gitlab.com/grchive/grchive-v3/shared/etl/connectors"
	"gitlab.com/grchive/grchive-v3/shared/etl/types"
	"golang.org/x/net/context"
	"io/ioutil"
	"net/http"
)

type EtlAWSConnectorUser struct {
	opts *EtlAWSOptions
}

func createAWSConnectorUser(opts *EtlAWSOptions) (*EtlAWSConnectorUser, error) {
	return &EtlAWSConnectorUser{
		opts: opts,
	}, nil
}

func (c *EtlAWSConnectorUser) GetUserListing() ([]*types.EtlUser, *connectors.EtlSourceInfo, error) {
	marker := ""
	ctx := context.Background()

	for {
		endpoint := fmt.Sprintf("%s/?Action=ListUsers&Version=2010-05-08&MaxItems=1000", iamBaseUrl)
		if marker != "" {
			endpoint = endpoint + fmt.Sprintf("&Marker=%s", marker)
		}

		req, err := http.NewRequestWithContext(
			ctx,
			"GET",
			endpoint,
			nil,
		)
		if err != nil {
			return nil, nil, err
		}

		resp, err := c.opts.Client.Do(req)
		if err != nil {
			return nil, nil, err
		}
		defer resp.Body.Close()

		bodyData, err := ioutil.ReadAll(resp.Body)
		if err != nil {
			return nil, nil, err
		}

		if resp.StatusCode != http.StatusOK {
			return nil, nil, errors.New("AWS User Listing API Error: " + string(bodyData))
		}

		break
	}

	return nil, nil, nil
}
