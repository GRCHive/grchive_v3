package auth_utility

type OAuthClientRegistration struct {
	ClientName   string   `json:"client_name"`
	RedirectUris []string `json:"redirect_uris"`
}

type OAuthClient struct {
	ClientId     string   `json:"client_id"`
	ClientSecret string   `json:"client_secret"`
	RedirectUris []string `json:"redirect_uris"`
}
