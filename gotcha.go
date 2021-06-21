package gotcha

import (
	"net/http"
	urlPkg "net/url"
)

func New(url string, method string, options *Options) (*http.Response, error) {
	defaultOptions := NewDefaultOptions()

	// set url
	u, err := urlPkg.Parse(url)
	if err != nil {
		return nil, err
	}
	options.Url = u

	// set method
	options.Method = method

	// merge options
	mergedOptions, err := defaultOptions.Extend(options)
	if err != nil {
		return nil, err
	}

	// do request
	res, err := mergedOptions.Adapter.DoRequest(mergedOptions)
	if err != nil {
		return nil, err
	}

	return res, nil
}

func Get(url string, options *Options) (*http.Response, error) {
	return New(url, "GET", options)
}

func Post(url string, options *Options) (*http.Response, error) {
	return New(url, "POST", options)
}

func Put(url string, options *Options) (*http.Response, error) {
	return New(url, "PUT", options)
}

func Patch(url string, options *Options) (*http.Response, error) {
	return New(url, "PATCH", options)
}

func Delete(url string, options *Options) (*http.Response, error) {
	return New(url, "DELETE", options)
}
