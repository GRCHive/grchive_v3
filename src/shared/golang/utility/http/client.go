package http_utility

import (
	"net/http"
)

type HeaderInjectionRoundTripper struct {
	headers map[string]string
	proxy   http.RoundTripper
}

func (t *HeaderInjectionRoundTripper) RoundTrip(req *http.Request) (*http.Response, error) {
	for k, v := range t.headers {
		req.Header.Add(k, v)
	}

	if t.proxy != nil {
		return t.proxy.RoundTrip(req)
	} else {
		return http.DefaultTransport.RoundTrip(req)
	}
}

func CreateHeaderInjectionClient(headers map[string]string, proxy http.RoundTripper) HttpClient {
	return &http.Client{
		Transport: &HeaderInjectionRoundTripper{headers: headers, proxy: proxy},
	}
}
