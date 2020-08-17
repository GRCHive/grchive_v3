package http_utility

import "net/http"

type HttpClient interface {
	Do(*http.Request) (*http.Response, error)
}
