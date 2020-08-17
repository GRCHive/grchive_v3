package http_utility

import (
	"golang.org/x/net/context"
	"golang.org/x/oauth2"
)

func CreateOAuth2AuthorizedClient(ts oauth2.TokenSource) HttpClient {
	return oauth2.NewClient(context.Background(), ts)
}
