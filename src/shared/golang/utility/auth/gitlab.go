package auth_utility

import (
	"gitlab.com/grchive/grchive-v3/shared/utility/http"
	"golang.org/x/net/context"
	"golang.org/x/oauth2"
	"golang.org/x/oauth2/gitlab"
)

func CreateGitlabOAuthConfig(clientId string, clientSecret string, redirectUrl string, scopes ...string) *oauth2.Config {
	return &oauth2.Config{
		ClientID:     clientId,
		ClientSecret: clientSecret,
		RedirectURL:  redirectUrl,
		Endpoint:     gitlab.Endpoint,
		Scopes:       scopes,
	}
}

func CreateGitlabOAuthTokenSource(config *oauth2.Config, code string) (oauth2.TokenSource, error) {
	ctx := context.Background()
	token, err := config.Exchange(ctx, code)
	if err != nil {
		return nil, err
	}

	return config.TokenSource(ctx, token), nil
}

func CreateGitlabHttpClient(ts oauth2.TokenSource) http_utility.HttpClient {
	return http_utility.CreateOAuth2AuthorizedClient(ts)
}
