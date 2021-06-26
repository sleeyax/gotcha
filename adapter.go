package gotcha

import (
	"net/http"
)

type Adapter interface {
	// DoRequest is a custom function that will be used by gotcha to make the actual request.
	DoRequest(options *Options) (*http.Response, error)
}

// RequestAdapter is a basic implementation of Adapter.
type RequestAdapter struct {
	// RoundTripper is a http.RoundTripper that will be used to do the request.
	//
	// Defaults to http.DefaultTransport.
	RoundTripper http.RoundTripper

	// Request is a function that builds the http.Request to send.
	//
	// Defaults to a function that derives Request the specified Options.
	Request func(*Options) *http.Request
}

func (ra *RequestAdapter) DoRequest(options *Options) (*http.Response, error) {
	if ra.Request == nil {
		ra.Request = func(o *Options) *http.Request {
			return &http.Request{
				Method: o.Method,
				URL:    o.FullUrl,
				Header: o.Headers,
				Body:   o.Body.Content,
			}
		}
	}

	if ra.RoundTripper == nil {
		ra.RoundTripper = http.DefaultTransport
	}

	req := ra.Request(options)

	if options.CookieJar != nil {
		for _, cookie := range options.CookieJar.Cookies(options.FullUrl) {
			req.AddCookie(cookie)
		}
	}

	res, err := ra.RoundTripper.RoundTrip(req)
	if err != nil {
		return nil, err
	}

	if options.CookieJar != nil {
		if rc := res.Cookies(); len(rc) > 0 {
			options.CookieJar.SetCookies(options.FullUrl, rc)
		}
	}

	return res, nil
}

// mockAdapter is only used for testing Adapter.
type mockAdapter struct {
	OnCalledDoRequest func(*Options) *http.Response
}

func (ma *mockAdapter) DoRequest(options *Options) (*http.Response, error) {
	return ma.OnCalledDoRequest(options), nil
}
