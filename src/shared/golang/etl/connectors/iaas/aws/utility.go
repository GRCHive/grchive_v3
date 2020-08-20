package aws

import (
	"encoding/xml"
	"errors"
	"fmt"
	"gitlab.com/grchive/grchive-v3/shared/etl/connectors"
	"gitlab.com/grchive/grchive-v3/shared/utility/http"
	"golang.org/x/net/context"
	"io/ioutil"
	"net/http"
	"reflect"
)

func awsGet(client http_utility.HttpClient, endpoint string, output interface{}) (*connectors.EtlSourceInfo, error) {
	ctx := context.Background()
	source := connectors.CreateSourceInfo()

	if reflect.TypeOf(output).Kind() != reflect.Ptr {
		return nil, errors.New("output must be a pointer to a slice.")
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
		return nil, errors.New("AWS API Error: " + string(bodyData))
	}

	err = xml.Unmarshal(bodyData, reflectOutPtr.Interface())
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

func awsPaginatedGet(client http_utility.HttpClient, resultName string, baseEndpoint string, output interface{}) (*connectors.EtlSourceInfo, error) {
	marker := ""
	source := connectors.CreateSourceInfo()

	if reflect.TypeOf(output).Kind() != reflect.Ptr {
		return nil, errors.New("output must be a pointer to a slice.")
	}

	reflectOutPtr := reflect.ValueOf(output)
	reflectOutSlice := reflectOutPtr.Elem()

	reflectBaseType := reflect.TypeOf(output).Elem().Elem()

	for {
		endpoint := baseEndpoint
		if marker != "" {
			endpoint = endpoint + fmt.Sprintf("&Marker=%s", marker)
		}

		responseBodyValue := reflect.New(reflectBaseType)
		cmdSrc, err := awsGet(client, endpoint, responseBodyValue.Interface())
		if err != nil {
			return nil, err
		}

		reflectOutSlice = reflect.Append(reflectOutSlice, responseBodyValue.Elem())
		source.MergeWith(cmdSrc)

		result := responseBodyValue.Elem().FieldByName(resultName)
		isTruncatedValue := result.FieldByName("IsTruncated").Interface().(bool)

		if !isTruncatedValue {
			break
		}

		marker = result.FieldByName("Marker").Interface().(string)
	}

	reflectOutPtr.Elem().Set(reflectOutSlice)
	return source, nil
}
