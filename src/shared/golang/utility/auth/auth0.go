package auth_utility

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"gitlab.com/grchive/grchive-v3/shared/utility/http"
	"golang.org/x/oauth2"
	"io/ioutil"
	"net/http"
	"net/url"
	"strings"
	"time"
)

type auth0TokenSource struct {
	Domain   string
	Audience string
	Client   *OAuthClient
}

type auth0Token struct {
	AccessToken string `json:"access_token"`
	ExpiresIn   int    `json:"expires_in"`
	Scope       string `json:"scope"`
	TokenType   string `json:"token_type"`
}

func (t *auth0TokenSource) Token() (*oauth2.Token, error) {
	endpoint := fmt.Sprintf("https://%s/oauth/token", t.Domain)

	params := url.Values{}
	params.Set("grant_type", "client_credentials")
	params.Set("client_id", t.Client.ClientId)
	params.Set("client_secret", t.Client.ClientSecret)
	params.Set("audience", t.Audience)

	resp, err := http.Post(endpoint, "application/x-www-form-urlencoded", strings.NewReader(params.Encode()))
	if err != nil {
		return nil, err
	}

	respBody, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	token := auth0Token{}
	err = json.Unmarshal(respBody, &token)
	if err != nil {
		return nil, err
	}

	return &oauth2.Token{
		AccessToken: token.AccessToken,
		TokenType:   token.TokenType,
		Expiry:      time.Now().Add(time.Duration(token.ExpiresIn) * time.Second),
	}, nil
}

func Auth0RegisterClient(domain string, reg OAuthClientRegistration) (*OAuthClient, error) {
	sendBody, err := json.Marshal(reg)
	if err != nil {
		return nil, err
	}

	resp, err := http.Post(fmt.Sprintf("https://%s/oidc/register", domain), "application/json", bytes.NewReader(sendBody))
	if err != nil {
		return nil, err
	}

	bodyData, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	if resp.StatusCode != http.StatusCreated {
		return nil, errors.New("Failed to register dynamic Auth0 client: " + string(bodyData))
	}

	client := OAuthClient{}
	err = json.Unmarshal(bodyData, &client)
	if err != nil {
		return nil, err
	}

	return &client, nil
}

func CreateAuth0OAuthTokenSource(domain string, audience string, client *OAuthClient) (oauth2.TokenSource, error) {
	return &auth0TokenSource{
		Domain:   domain,
		Client:   client,
		Audience: audience,
	}, nil
}

func CreateAuth0HttpClient(ts oauth2.TokenSource) http_utility.HttpClient {
	return http_utility.CreateOAuth2AuthorizedClient(ts)
}
