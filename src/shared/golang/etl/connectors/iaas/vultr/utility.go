package vultr

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

func vultrGet(client http_utility.HttpClient, endpoint string, output interface{}) (*connectors.EtlSourceInfo, error) {
	ctx := context.Background()
	source := connectors.CreateSourceInfo()

	if reflect.TypeOf(output).Kind() != reflect.Ptr {
		return nil, errors.New("output must be a pointer.")
	}

	reflectOutPtr := reflect.ValueOf(output)

	req, err := http.NewRequestWithContext(
		ctx,
		"GET",
		endpoint,
		nil,
	)
	if err != nil {
		return nil, err
	}

	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	bodyData, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	if resp.StatusCode != http.StatusOK {
		return nil, errors.New("Vultr API Error: " + string(bodyData))
	}

	err = json.Unmarshal(bodyData, reflectOutPtr.Interface())
	if err != nil {
		return nil, err
	}

	cmd := &connectors.EtlCommandInfo{
		Command: endpoint,
		RawData: string(bodyData),
	}
	source.AddCommand(cmd)
	return source, nil
}

func vultrPaginatedGet(client http_utility.HttpClient, baseEndpoint string, output interface{}) (*connectors.EtlSourceInfo, error) {
	cursor := ""
	source := connectors.CreateSourceInfo()

	if reflect.TypeOf(output).Kind() != reflect.Ptr {
		return nil, errors.New("output must be a pointer to a slice.")
	}

	reflectOutPtr := reflect.ValueOf(output)
	reflectOutSlice := reflectOutPtr.Elem()

	reflectBaseType := reflect.TypeOf(output).Elem().Elem()

	for {
		endpoint := baseEndpoint
		if cursor != "" {
			endpoint = endpoint + fmt.Sprintf("?cursor=%s", cursor)
		}

		responseBodyValue := reflect.New(reflectBaseType)
		cmdSrc, err := vultrGet(client, endpoint, responseBodyValue.Interface())
		if err != nil {
			return nil, err
		}

		reflectOutSlice = reflect.Append(reflectOutSlice, responseBodyValue.Elem())
		source.MergeWith(cmdSrc)

		result := responseBodyValue.Elem()
		meta := result.FieldByName("Meta").Interface().(vultrMeta)

		if meta.Links.Next == "" {
			break
		}

		cursor = meta.Links.Next
	}

	reflectOutPtr.Elem().Set(reflectOutSlice)
	return source, nil
}
