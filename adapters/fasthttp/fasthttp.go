package fasthttp

import (
	"bytes"
	"encoding/base64"
	"fmt"
	"github.com/sleeyax/gotcha"
	"github.com/valyala/fasthttp"
	"github.com/valyala/fasthttp/fasthttpadaptor"
	"github.com/valyala/fasthttp/fasthttpproxy"
	"io/ioutil"
	"net/http"
	"strings"
)

type Adapter struct{}

func NewAdapter() *Adapter {
	return &Adapter{}
}

func (a *Adapter) DoRequest(options *gotcha.Options) (*gotcha.Response, error) {
	req := fasthttp.AcquireRequest()
	defer fasthttp.ReleaseRequest(req)
	res := fasthttp.AcquireResponse()
	defer fasthttp.ReleaseResponse(res)

	req.Header.SetMethod(options.Method)
	req.Header.SetRequestURI(options.FullUrl.RequestURI())

	for key, values := range options.Headers {
		for _, value := range values {
			req.Header.Add(key, value)
		}
	}

	var c *fasthttp.Client

	if options.Proxy != nil {
		proxy := options.Proxy.RequestURI()
		c.Dial = fasthttpproxy.FasthttpHTTPDialer(proxy)
		if v := strings.Split(proxy, ":"); len(v) >= 4 {
			var auth = "Basic "
			if len(v) == 5 {
				auth += base64.StdEncoding.EncodeToString([]byte(fmt.Sprintf("%s:%s", v[1], v[2])))
			} else {
				auth += base64.StdEncoding.EncodeToString([]byte(fmt.Sprintf("%s:%s", v[0], v[1])))
			}

			req.Header.Set("Proxy-Authorization", auth)
		}
	}

	if options.CookieJar != nil {
		var cookies []*http.Cookie
		res.Header.VisitAllCookie(func(key, value []byte) {
			sk := string(key)
			sv := string(value)
			cookies = append(cookies, &http.Cookie{
				Name:       sk,
				Value:      sv,
				Domain:     string(req.RequestURI()),
			})
		})

		options.CookieJar.SetCookies(options.FullUrl, cookies)
	}

	if err := c.Do(req, res); err != nil {
		return nil, err
	}

	reqCtx := &fasthttp.RequestCtx{}

	req.CopyTo(&reqCtx.Request)
	res.CopyTo(&reqCtx.Response)

	r, err := toResponse(reqCtx)
	if err != nil {
		return nil, err
	}

	return &gotcha.Response{r, options.UnmarshalJson}, nil
}

func toResponse(ctx *fasthttp.RequestCtx) (*http.Response, error) {
	var r *http.Request

	if err := fasthttpadaptor.ConvertRequest(ctx, r, false); err != nil {
		return nil, err
	}

	resp := &http.Response{
		Status:           statusMap[ctx.Response.StatusCode()],
		StatusCode:       ctx.Response.StatusCode(),
		Proto:            "HTTP/1.1",
		ProtoMajor:       1,
		ProtoMinor:       1,
		Body:             ioutil.NopCloser(bytes.NewBuffer(ctx.Response.Body())),
		ContentLength:    int64(len(ctx.Response.Body())),
		Uncompressed:     false,
		Request:          r,
		TLS:              ctx.TLSConnectionState(),
	}

	ctx.Response.Header.VisitAll(func(k, v []byte) {
		sk := string(k)
		sv := string(v)

		switch sk {
		case "Transfer-Encoding":
			resp.TransferEncoding = append(resp.TransferEncoding, sv)
		default:
			r.Header.Set(sk, sv)
		}

	})

	return resp, nil
}
