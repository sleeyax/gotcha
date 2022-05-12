package fhttp

import (
	"github.com/sleeyax/gotcha"
	fhttp "github.com/useflyent/fhttp"
	"net/http"
)

type Adapter struct {
	// Optional fhttp Transport options.
	Transport *fhttp.Transport
}

func NewAdapter() *Adapter {
	return &Adapter{}
}

func (a *Adapter) DoRequest(options *gotcha.Options) (*gotcha.Response, error) {
	req := &fhttp.Request{
		Method: options.Method,
		URL:    options.FullUrl,
		Header: fhttp.Header(options.Headers.Clone()),
		Body:   options.Body,
	}

	if a.Transport == nil {
		a.Transport = fhttp.DefaultTransport.(*fhttp.Transport)
	}

	if options.Proxy != nil {
		a.Transport.Proxy = fhttp.ProxyURL(options.Proxy)
	}

	if options.CookieJar != nil {
		for _, cookie := range options.CookieJar.Cookies(options.FullUrl) {
			req.AddCookie(&fhttp.Cookie{Name: cookie.Name, Value: cookie.Value})
		}
	}

	res, err := a.Transport.RoundTrip(req)
	if err != nil {
		return nil, err
	}

	r := toResponse(req, res)

	if options.CookieJar != nil {
		if rc := r.Cookies(); len(rc) > 0 {
			options.CookieJar.SetCookies(options.FullUrl, rc)
		}
	}

	return &gotcha.Response{r, options.UnmarshalJson}, nil
}

// toResponse converts fhttp response to an original http response.
func toResponse(req *fhttp.Request, res *fhttp.Response) *http.Response {
	return &http.Response{
		Status:           res.Status,
		StatusCode:       res.StatusCode,
		Proto:            res.Proto,
		ProtoMajor:       res.ProtoMajor,
		ProtoMinor:       res.ProtoMinor,
		Header:           http.Header(res.Header),
		Body:             res.Body,
		ContentLength:    res.ContentLength,
		TransferEncoding: res.TransferEncoding,
		Close:            res.Close,
		Uncompressed:     res.Uncompressed,
		Trailer:          http.Header(res.Trailer),
		Request:          toRequest(req),
		TLS:              res.TLS,
	}
}

// toResponse converts a fhttp request to an original http response.
func toRequest(req *fhttp.Request) *http.Request {
	return &http.Request{
		Method: req.Method,
		URL:    req.URL,
		Header: http.Header(req.Header),
		Body:   req.Body,
	}
}
