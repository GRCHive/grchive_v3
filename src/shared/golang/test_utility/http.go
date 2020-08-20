package test_utility

import (
	"io/ioutil"
	"net/http"
	"strings"
)

func WrapHttpResponse(data string) *http.Response {
	body := ioutil.NopCloser(strings.NewReader(data))
	return &http.Response{
		StatusCode: http.StatusOK,
		Body:       body,
	}
}
