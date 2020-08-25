package auth_utility

import (
	"fmt"
	"gitlab.com/grchive/grchive-v3/shared/utility/http"
)

func CreateCloudflareHttpClient(token string) http_utility.HttpClient {
	return http_utility.CreateHeaderInjectionClient(map[string]string{
		"Authorization": fmt.Sprintf("Bearer %s", token),
	}, nil)
}
