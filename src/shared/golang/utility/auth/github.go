package auth_utility

import (
	"fmt"
	"github.com/lestrrat-go/jwx/jwa"
	"github.com/lestrrat-go/jwx/jwt"
	"gitlab.com/grchive/grchive-v3/shared/utility/http"
	"gitlab.com/grchive/grchive-v3/shared/utility/time"
	"time"
)

func CreateGithubJWTToken(clock time_utility.Clock, appId string, keyFname string) (string, error) {
	privateKey, err := ReadRSAPrivateKeyFromPEM(keyFname)
	if err != nil {
		return "", err
	}

	issued := clock.Now()
	expiration := issued.Add(10 * time.Minute)

	token := jwt.New()
	token.Set(jwt.IssuedAtKey, issued)
	token.Set(jwt.ExpirationKey, expiration)
	token.Set(jwt.IssuerKey, appId)

	signedToken, err := jwt.Sign(token, jwa.RS256, privateKey)
	if err != nil {
		return "", err
	}

	return string(signedToken), nil
}

func CreateGithubHttpJWTClient(jwt string) http_utility.HttpClient {
	return http_utility.CreateHeaderInjectionClient(map[string]string{
		"Authorization": fmt.Sprintf("Bearer %s", jwt),
		"Accept":        "application/vnd.github.machine-man-preview+json",
	}, nil)
}

func CreateGithubHttpInstallationClient(token string) http_utility.HttpClient {
	return http_utility.CreateHeaderInjectionClient(map[string]string{
		"Authorization": fmt.Sprintf("token %s", token),
		"Accept":        "application/vnd.github.machine-man-preview+json",
	}, nil)
}
