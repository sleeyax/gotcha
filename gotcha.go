// Package gotcha is an advanced HTTP client with a got-like API.
// It can interface with other HTTP packages through an adapter.
package gotcha

import "net/http"

func DoRequest(url string, method string, options ...*Options) (*Response, error) {
	client, err := NewClient(&Options{})
	if err != nil {
		return nil, err
	}
	return client.DoRequest(method, url, options...)
}

func Get(url string, options ...*Options) (*Response, error) {
	return DoRequest(url, http.MethodGet, options...)
}

func Post(url string, options ...*Options) (*Response, error) {
	return DoRequest(url, http.MethodPost, options...)
}

func Put(url string, options ...*Options) (*Response, error) {
	return DoRequest(url, http.MethodPut, options...)
}

func Patch(url string, options ...*Options) (*Response, error) {
	return DoRequest(url, http.MethodPatch, options...)
}

func Delete(url string, options ...*Options) (*Response, error) {
	return DoRequest(url, http.MethodDelete, options...)
}
