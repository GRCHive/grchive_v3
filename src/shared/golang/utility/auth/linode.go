package auth_utility

import (
	"gitlab.com/grchive/grchive-v3/shared/utility/http"
	"golang.org/x/net/context"
	"golang.org/x/oauth2"
)

func CreateLinodeOAuthConfig(clientId string, clientSecret string, redirectUrl string, scopes ...string) *oauth2.Config {
	return &oauth2.Config{
		ClientID:     clientId,
		ClientSecret: clientSecret,
		RedirectURL:  redirectUrl,
		Endpoint: oauth2.Endpoint{
			AuthURL:  "https://login.linode.com/oauth/authorize",
			TokenURL: "https://login.linode.com/oauth/token",
		},
		Scopes: scopes,
	}
}

func CreateLinodeOAuthTokenSource(config *oauth2.Config, code string) (oauth2.TokenSource, error) {
	ctx := context.Background()
	token, err := config.Exchange(ctx, code)
	if err != nil {
		return nil, err
	}

	return config.TokenSource(ctx, token), nil
}

func CreateLinodeHttpClient(ts oauth2.TokenSource) http_utility.HttpClient {
	return http_utility.CreateOAuth2AuthorizedClient(ts)
}
