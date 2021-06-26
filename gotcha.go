package gotcha

import (
	"net/http"
)

func DoRequest(url string, method string, options ...*Options) (*http.Response, error) {
	client, err := NewClient(&Options{})
	if err != nil {
		return nil, err
	}
	return client.DoRequest(method, url, options...)
}

func Get(url string, options ...*Options) (*http.Response, error) {
	return DoRequest(url, "GET", options...)
}

func Post(url string, options ...*Options) (*http.Response, error) {
	return DoRequest(url, "POST", options...)
}

func Put(url string, options ...*Options) (*http.Response, error) {
	return DoRequest(url, "PUT", options...)
}

func Patch(url string, options ...*Options) (*http.Response, error) {
	return DoRequest(url, "PATCH", options...)
}

func Delete(url string, options ...*Options) (*http.Response, error) {
	return DoRequest(url, "DELETE", options...)
}
