package auth_utility

import (
	"fmt"
	"gitlab.com/grchive/grchive-v3/shared/utility/http"
	"golang.org/x/net/context"
	"golang.org/x/oauth2"
	"net/url"
	"strings"
)

const AzureGraphResource = "graph.microsoft.com"
const AzureManagementResource = "management.core.windows.net"

type AzureOAuthSetup struct {
	Tenant       string
	ClientId     string
	ClientSecret string
	RedirectURL  string
	Scopes       []string
}

func (s AzureOAuthSetup) ToOAuthConfig() *oauth2.Config {
	return &oauth2.Config{
		ClientID:     s.ClientId,
		ClientSecret: s.ClientSecret,
		RedirectURL:  s.RedirectURL,
		Endpoint: oauth2.Endpoint{
			AuthURL:  fmt.Sprintf("https://login.microsoftonline.com/%s/oauth2/v2.0/authorize", s.Tenant),
			TokenURL: fmt.Sprintf("https://login.microsoftonline.com/%s/oauth2/v2.0/token", s.Tenant),
		},
		Scopes: s.Scopes,
	}
}

// We need to create a separate token source for Microsoft Graph and Microsoft Azure Management API.
// We should be able to split the resource just by using the host of the URL.
func CreateAzureOAuthTokenSource(config *oauth2.Config, code string) (map[string]oauth2.TokenSource, error) {
	ctx := context.Background()
	ret := map[string]oauth2.TokenSource{}

	for _, scope := range config.Scopes {
		// Assume that the only scopes we need to get a token source for are the microsoft ones...
		if !strings.HasPrefix(scope, "https://") {
			continue
		}

		scopeUrl, err := url.Parse(scope)
		if err != nil {
			return nil, err
		}

		// If the resource already has a token source we can ignore.
		_, ok := ret[scopeUrl.Host]
		if ok {
			continue
		}

		token, err := config.Exchange(
			ctx,
			code,
			oauth2.SetAuthURLParam("scope", scope),
		)
		if err != nil {
			return nil, err
		}
		ret[scopeUrl.Host] = config.TokenSource(ctx, token)
	}
	return ret, nil
}

func CreateAzureHttpClient(ts oauth2.TokenSource) http_utility.HttpClient {
	return http_utility.CreateOAuth2AuthorizedClient(ts)
}
