package graphql_utility

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"gitlab.com/grchive/grchive-v3/shared/utility/http"
	"io/ioutil"
	"net/http"
)

type GraphQLRequestBody struct {
	Query     string                 `json:"query"`
	Variables map[string]interface{} `json:"variables"`
}

// Returns the raw output as well as any errors.
// The output is marshaled into resp if it's not nil.
func SendGraphQLRequest(endpoint string, client http_utility.HttpClient, request GraphQLRequestBody, resp interface{}) (string, error) {
	ctx := context.Background()

	requestBuffer := bytes.Buffer{}
	err := json.NewEncoder(&requestBuffer).Encode(request)
	if err != nil {
		return "", err
	}

	httpRequest, err := http.NewRequestWithContext(
		ctx,
		"POST",
		endpoint,
		&requestBuffer,
	)
	if err != nil {
		return "", err
	}

	httpResponse, err := client.Do(httpRequest)
	if err != nil {
		return "", err
	}
	defer httpResponse.Body.Close()

	bodyData, err := ioutil.ReadAll(httpResponse.Body)
	if err != nil {
		return "", err
	}
	rawBodyData := string(bodyData)

	if httpResponse.StatusCode != http.StatusOK {
		return "", errors.New("GraphQL error: " + rawBodyData)
	}

	if resp != nil {
		err = json.Unmarshal(bodyData, resp)
		if err != nil {
			return "", err
		}
	}

	return rawBodyData, nil
}
