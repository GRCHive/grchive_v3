package heroku

import (
	"encoding/json"
	"errors"
	"gitlab.com/grchive/grchive-v3/shared/etl/connectors"
	"gitlab.com/grchive/grchive-v3/shared/utility/http"
	"golang.org/x/net/context"
	"io/ioutil"
	"net/http"
	"reflect"
)

func herokuGet(client http_utility.HttpClient, endpoint string, output interface{}, addtlHeaders http.Header) (*http.Response, *connectors.EtlSourceInfo, error) {
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
		return nil, nil, errors.New("Heroku API Error: " + string(bodyData))
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

func herokuPaginatedGet(client http_utility.HttpClient, baseEndpoint string, output interface{}) (*connectors.EtlSourceInfo, error) {
	nextRange := ""
	source := connectors.CreateSourceInfo()

	if reflect.TypeOf(output).Kind() != reflect.Ptr {
		return nil, errors.New("output must be a pointer to a slice.")
	}

	reflectOutPtr := reflect.ValueOf(output)
	reflectOutSlice := reflectOutPtr.Elem()

	reflectBaseType := reflect.TypeOf(output).Elem().Elem()

	for {
		endpoint := baseEndpoint
		addtlHeaders := http.Header{}
		if nextRange != "" {
			addtlHeaders.Set("Range", nextRange)
		}

		responseBodyValue := reflect.New(reflectBaseType)
		resp, cmdSrc, err := herokuGet(client, endpoint, responseBodyValue.Interface(), addtlHeaders)
		if err != nil {
			return nil, err
		}

		reflectOutSlice = reflect.Append(reflectOutSlice, responseBodyValue.Elem())
		source.MergeWith(cmdSrc)

		if resp.StatusCode != http.StatusPartialContent {
			break
		}

		nextRange = resp.Header.Get("Next-Range")
	}

	reflectOutPtr.Elem().Set(reflectOutSlice)
	return source, nil
}
