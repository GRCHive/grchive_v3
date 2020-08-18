package github

import (
	"encoding/json"
	"errors"
	"fmt"
	"gitlab.com/grchive/grchive-v3/shared/etl/connectors"
	"gitlab.com/grchive/grchive-v3/shared/utility/http"
	"golang.org/x/net/context"
	"io/ioutil"
	"net/http"
)

type EtlGithubOptions struct {
	Client http_utility.HttpClient
	OrgId  string
}

type EtlGithubConnector struct {
	opts  *EtlGithubOptions
	users *EtlGithubConnectorUser
}

const baseUrl string = "https://api.github.com"
const graphqlEndpoint string = "https://api.github.com/graphql"

func (c *EtlGithubConnector) GetUserInterface() (connectors.EtlConnectorUserInterface, error) {
	return c.users, nil
}

func CreateGithubConnector(opts *EtlGithubOptions) (*EtlGithubConnector, error) {
	var err error
	ret := EtlGithubConnector{
		opts: opts,
	}
	ret.users, err = createGithubConnectorUser(opts)

	if err != nil {
		return nil, err
	}

	return &ret, nil
}

func (c *EtlGithubConnector) GetInstallationAccessToken(installation string) (string, error) {
	ctx := context.Background()
	endpoint := fmt.Sprintf(
		"%s/app/installations/%s/access_tokens",
		baseUrl,
		installation,
	)

	req, err := http.NewRequestWithContext(
		ctx,
		"POST",
		endpoint,
		nil,
	)
	if err != nil {
		return "", err
	}

	resp, err := c.opts.Client.Do(req)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	bodyData, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return "", err
	}

	if resp.StatusCode != http.StatusCreated {
		return "", errors.New("Github Get Installation Access Token:" + string(bodyData) + " :: " + endpoint)
	}

	type ResponseBody struct {
		Token string `json:"token"`
	}

	body := ResponseBody{}
	err = json.Unmarshal(bodyData, &body)
	if err != nil {
		return "", err
	}

	return body.Token, nil
}
