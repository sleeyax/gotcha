package gotcha

import (
	"net/http"
)

func DoRequest(url string, method string, options ...*Options) (*http.Response, error) {
	o := &Options{}
	var err error

	for _, option := range options {
		o, err = o.Extend(option)
	}

	client, err := NewClient(o)
	if err != nil {
		return nil, err
	}
	return client.DoRequest(url, method)
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
