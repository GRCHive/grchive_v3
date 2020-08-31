package auth0

import (
	"encoding/json"
	"errors"
	"fmt"
	"gitlab.com/grchive/grchive-v3/shared/etl/connectors"
	"gitlab.com/grchive/grchive-v3/shared/utility/http"
	"golang.org/x/net/context"
	"io/ioutil"
	"net/http"
	"reflect"
)

func auth0Get(client http_utility.HttpClient, endpoint string, output interface{}) (*http.Response, *connectors.EtlSourceInfo, error) {
	ctx := context.Background()
	source := connectors.CreateSourceInfo()

	if reflect.TypeOf(output).Kind() != reflect.Ptr {
		return nil, nil, errors.New("output must be a pointer.")
	}

	reflectOutPtr := reflect.ValueOf(output)

	req, err := http.NewRequestWithContext(
		ctx,
		"GET",
		endpoint,
		nil,
	)
	if err != nil {
		return nil, nil, err
	}

	resp, err := client.Do(req)
	if err != nil {
		return nil, nil, err
	}
	defer resp.Body.Close()

	bodyData, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return nil, nil, err
	}

	if resp.StatusCode >= http.StatusBadRequest {
		return nil, nil, errors.New("Auth0 API Error: " + string(bodyData))
	}

	err = json.Unmarshal(bodyData, reflectOutPtr.Interface())
	if err != nil {
		return nil, nil, err
	}

	cmd := &connectors.EtlCommandInfo{
		Command: endpoint,
		RawData: string(bodyData),
	}
	source.AddCommand(cmd)
	return resp, source, nil
}

func auth0PaginatedGet(client http_utility.HttpClient, baseEndpoint string, output interface{}) (*connectors.EtlSourceInfo, error) {
	source := connectors.CreateSourceInfo()

	if reflect.TypeOf(output).Kind() != reflect.Ptr {
		return nil, errors.New("output must be a pointer to a slice.")
	}

	reflectOutPtr := reflect.ValueOf(output)
	reflectOutSlice := reflectOutPtr.Elem()

	reflectBaseType := reflect.TypeOf(output).Elem().Elem()

	page := 0
	perPage := 50

	for {
		endpoint := fmt.Sprintf("%s?page=%d&per_page=%d", baseEndpoint, page, perPage)

		responseBodyValue := reflect.New(reflectBaseType)
		_, cmdSrc, err := auth0Get(client, endpoint, responseBodyValue.Interface())
		if err != nil {
			return nil, err
		}

		newPage := responseBodyValue.Elem()
		if newPage.Len() == 0 {
			break
		}

		reflectOutSlice = reflect.Append(reflectOutSlice, newPage)
		source.MergeWith(cmdSrc)

		page += 1
	}

	reflectOutPtr.Elem().Set(reflectOutSlice)
	return source, nil
}
