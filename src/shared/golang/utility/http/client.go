package http_utility

import (
	"net/http"
)

type HeaderInjectionRoundTripper struct {
	headers map[string]string
}

func (t *HeaderInjectionRoundTripper) RoundTrip(req *http.Request) (*http.Response, error) {
	for k, v := range t.headers {
		req.Header.Add(k, v)
	}
	return http.DefaultTransport.RoundTrip(req)
}

func CreateHeaderInjectionClient(headers map[string]string) HttpClient {
	return &http.Client{
		Transport: &HeaderInjectionRoundTripper{headers: headers},
	}
}
