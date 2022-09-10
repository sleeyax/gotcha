package gotcha

import (
	bytesPkg "bytes"
	"github.com/Sleeyax/urlValues"
	"github.com/sleeyax/gotcha/internal/utils"
	"io"
	"net/http"
	"strconv"
	"strings"
	"time"
)

type Client struct {
	// Instance-specific configuration options.
	Options *Options
}

// NewClient creates a new HTTP client based on default Options which can be extended.
func NewClient(options *Options) (*Client, error) {
	client := &Client{NewDefaultOptions()}
	return client.Extend(options)
}

// Extend returns a new Client with merged Options.
func (c *Client) Extend(options *Options) (*Client, error) {
	opts, err := c.Options.Extend(options)
	if err != nil {
		return nil, err
	}
	return &Client{opts}, nil
}

func (c *Client) DoRequest(method string, url string, options ...*Options) (*Response, error) {
	for _, hook := range c.Options.Hooks.Init {
		hook(c.Options)
	}

	o := &(*c.Options)

	for _, option := range options {
		var err error
		o, err = o.Extend(option)
		if err != nil {
			return nil, err
		}
	}

	u, err := utils.MergeUrl(o.PrefixURL, url, false)
	if err != nil {
		return nil, err
	}

	o.FullUrl = u
	o.Method = method
	o.URI = url

	if sp := o.SearchParams; len(sp) != 0 {
		if _, ok := sp[urlValues.OrderKey]; ok {
			o.FullUrl.RawQuery = sp.EncodeWithOrder()
		} else {
			o.FullUrl.RawQuery = sp.Encode()
		}
	}

	c.ParseBody(o)
	defer c.CloseBody(o)

	retry := func(res *Response, err error) (*Response, error) {
		for _, hook := range o.Hooks.BeforeRetry {
			hook(o, err, o.retries)
		}
		timeout, e := c.getTimeout(o, res)
		if e != nil {
			return nil, e
		}
		timeout = o.RetryOptions.CalculateTimeout(o.retries, o.RetryOptions, timeout, err)
		time.Sleep(timeout)
		o.retries++
		return c.DoRequest(method, url, o)
	}

	for _, hook := range o.Hooks.BeforeRequest {
		hook(o)
	}

	res, err := o.Adapter.DoRequest(o)

	if err == nil {
		for _, hook := range o.Hooks.AfterResponse {
			var retryFunc RetryFunc = func(options *Options) (*Response, error) {
				o, err = o.Extend(options)
				if err != nil {
					return nil, err
				}
				return retry(res, nil)
			}
			r, e := hook(res, retryFunc)
			if e != nil {
				return nil, e
			}
			if r != nil {
				res = r
			}
		}
	}

	if err != nil && (!o.Retry || (o.retries >= o.RetryOptions.Limit)) {
		return nil, err
	}

	if o.Retry {
		if o.retries >= o.RetryOptions.Limit {
			return res, NewMaxRetriesExceededError()
		}

		if (err != nil && utils.StringArrayContains(o.RetryOptions.ErrorCodes, err.Error())) || (utils.IntArrayContains(o.RetryOptions.StatusCodes, res.StatusCode) && utils.StringArrayContains(o.RetryOptions.Methods, method)) {
			return retry(res, err)
		}
	}

	if o.FollowRedirect && res.Header.Get("location") != "" && utils.IntArrayContains(RedirectStatusCodes, res.StatusCode) {
		// we don't care about the response since we're redirecting
		res.Body.Close()

		if o.RedirectOptions.Limit != 0 && len(o.redirectUrls) >= o.RedirectOptions.Limit {
			return res, NewMaxRedirectsExceededError(len(o.redirectUrls))
		}

		if o.RedirectOptions.RewriteMethods || (res.StatusCode == 303 && o.Method != http.MethodGet && o.Method != http.MethodHead) {
			o.Method = http.MethodGet
			c.CloseBody(o)
			o.Headers.Del("content-length")
			o.Headers.Del("content-type")
		}

		currentUrl := o.FullUrl
		redirectUrl, err := utils.MergeUrl(currentUrl.String(), res.Header.Get("location"), true)
		if err != nil {
			return nil, err
		}

		// remove sensitive headers when redirecting to a different domain
		if redirectUrl.Hostname() != currentUrl.Hostname() || redirectUrl.Port() != currentUrl.Port() {
			o.Headers.Del("host")
			o.Headers.Del("cookie")
			o.Headers.Del("authorization")
		}

		c.updateRequestCookies(o, res)
		o.PrefixURL = ""
		o.redirectUrls = append(o.redirectUrls, redirectUrl)

		for _, hook := range o.Hooks.BeforeRedirect {
			hook(o, res)
		}

		return c.DoRequest(o.Method, redirectUrl.String(), o)
	}

	return res, nil
}

// Do is an alias of DoRequest.
func (c *Client) Do(method string, url string, options ...*Options) (*Response, error) {
	return c.DoRequest(method, url, options...)
}

func (c *Client) getTimeout(o *Options, response *Response) (time.Duration, error) {
	if o.RetryOptions.RetryAfter == false {
		return 0, nil
	}

	retryAfter := strings.TrimSpace(response.Header.Get("retry-after"))

	// Response header doesn't specify timeout, so default to 0.
	// Note that the user can still overwrite this behaviour in the configuration RetryOptions.
	if retryAfter == "" {
		return 0, nil
	}

	// retryAfter is <delay-seconds>
	if delay, err := strconv.Atoi(retryAfter); err == nil {
		return time.Second * time.Duration(delay), nil
	}

	// retryAfter is <http-date>
	dateTime, err := http.ParseTime(retryAfter)
	if err == nil {
		return dateTime.Sub(time.Now()), nil
	}
	return 0, err
}

// updateRequestCookies synchronizes request cookies that were manually set through the Cookie http.Header
// with corresponding cookies from a http.Response after redirect.
//
// If CookieJar is present and there was some initial cookies provided
// via the request header, then we may need to alter the initial
// cookies as we follow redirects since each redirect may end up
// modifying a pre-existing cookie.
//
// Since cookies already set in the request header do not contain
// information about the original domain and path, the logic below
// assumes any new set cookies override the original cookie
// regardless of domain or path.
func (c *Client) updateRequestCookies(o *Options, response *Response) {
	cookies := c.getRequestCookies(o)
	if o.CookieJar != nil && cookies != nil {
		// changed denotes whether or not a response cookie has a different value than a request cookie
		var changed bool
		for _, cookie := range response.Cookies() {
			if _, ok := cookies[cookie.Name]; ok {
				delete(cookies, cookie.Name)
				changed = true
			}
		}
		if changed {
			o.Headers.Del("Cookie")
			var cks []string
			for _, cs := range cookies {
				for _, cookie := range cs {
					cks = append(cks, cookie.Name+"="+cookie.Value)
				}
			}
			o.Headers.Set("Cookie", strings.Join(cks, "; "))
		}
	}
}

// getRequestCookies returns the cookies that were manually set in Options.Headers.
func (c *Client) getRequestCookies(o *Options) map[string][]*http.Cookie {
	var cookies map[string][]*http.Cookie

	if o.CookieJar != nil && o.Headers.Get("cookie") != "" {
		cookies = make(map[string][]*http.Cookie)
		req := http.Request{Header: o.Headers}
		for _, c := range req.Cookies() {
			cookies[c.Name] = append(cookies[c.Name], c)
		}
	}

	return cookies
}

// CloseBody clears the Body, Form and Json fields.
func (c *Client) CloseBody(o *Options) {
	if o.Body != nil {
		o.Body.Close()
	}
	o.Body = nil
	o.Form = nil
	o.Json = nil
}

// ParseBody parses Form or Json (in that order) into Body.
func (c *Client) ParseBody(o *Options) error {
	if len(o.Form) != 0 {
		encoded := o.Form.EncodeWithOrder()
		o.Body = io.NopCloser(strings.NewReader(encoded))
		return nil
	} else if j := o.Json; len(j) != 0 {
		bytes, err := o.MarshalJson(j)
		if err != nil {
			return err
		}
		o.Body = io.NopCloser(bytesPkg.NewReader(bytes))
		return nil
	}
	return nil
}

func (c *Client) Get(url string, options ...*Options) (*Response, error) {
	return c.DoRequest(http.MethodGet, url, options...)
}

func (c *Client) Post(url string, options ...*Options) (*Response, error) {
	return c.DoRequest(http.MethodPost, url, options...)
}

func (c *Client) Put(url string, options ...*Options) (*Response, error) {
	return c.DoRequest(http.MethodPut, url, options...)
}

func (c *Client) Patch(url string, options ...*Options) (*Response, error) {
	return c.DoRequest(http.MethodPatch, url, options...)
}

func (c *Client) Delete(url string, options ...*Options) (*Response, error) {
	return c.DoRequest(http.MethodDelete, url, options...)
}

func (c *Client) Head(url string, options ...*Options) (*Response, error) {
	return c.DoRequest(http.MethodHead, url, options...)
}
