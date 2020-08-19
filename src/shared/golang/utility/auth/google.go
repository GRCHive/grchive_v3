package auth_utility

import (
	"gitlab.com/grchive/grchive-v3/shared/utility/http"
	"golang.org/x/net/context"
	"golang.org/x/oauth2"
	"golang.org/x/oauth2/google"
	"io/ioutil"
)

func CreateGoogleOAuthTokenSource(jwtFname string, subject string, scopes ...string) (oauth2.TokenSource, error) {
	credentials, err := ioutil.ReadFile(jwtFname)
	if err != nil {
		return nil, err
	}

	config, err := google.JWTConfigFromJSON(credentials, scopes...)
	if err != nil {
		return nil, err
	}

	if subject != "" {
		config.Subject = subject
	}

	ts := config.TokenSource(context.Background())
	return ts, nil
}

func CreateGoogleHttpClient(ts oauth2.TokenSource) http_utility.HttpClient {
	return http_utility.CreateOAuth2AuthorizedClient(ts)
}
