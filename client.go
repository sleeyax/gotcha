package gotcha

import (
	bytesPkg "bytes"
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

	for _, option := range options {
		var err error
		c.Options, err = c.Options.Extend(option)
		if err != nil {
			return nil, err
		}
	}

	u, err := utils.MergeUrl(c.Options.PrefixURL, url, false)
	if err != nil {
		return nil, err
	}

	c.Options.FullUrl = u
	c.Options.Method = method
	c.Options.URI = url

	if sp := c.Options.SearchParams; len(sp) != 0 {
		c.Options.FullUrl.RawQuery = sp.EncodeWithOrder()
	}

	c.ParseBody()
	defer c.CloseBody()

	retry := func(res *Response, err error) (*Response, error) {
		for _, hook := range c.Options.Hooks.BeforeRetry {
			hook(c.Options, err, c.Options.retries)
		}
		timeout, e := c.getTimeout(res)
		if e != nil {
			return nil, e
		}
		timeout = c.Options.RetryOptions.CalculateTimeout(c.Options.retries, c.Options.RetryOptions, timeout, err)
		time.Sleep(timeout)
		c.Options.retries++
		return c.DoRequest(method, url)
	}

	for _, hook := range c.Options.Hooks.BeforeRequest {
		hook(c.Options)
	}

	res, err := c.Options.Adapter.DoRequest(c.Options)

	if err == nil {
		for _, hook := range c.Options.Hooks.AfterResponse {
			var retryFunc RetryFunc = func(o *Options) (*Response, error) {
				c.Options, err = c.Options.Extend(o)
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

	if err != nil && (!c.Options.Retry || (c.Options.retries >= c.Options.RetryOptions.Limit)) {
		return nil, err
	}

	if c.Options.Retry {
		if c.Options.retries >= c.Options.RetryOptions.Limit {
			return res, NewMaxRetriesExceededError()
		}

		if (err != nil && utils.StringArrayContains(c.Options.RetryOptions.ErrorCodes, err.Error())) || (utils.IntArrayContains(c.Options.RetryOptions.StatusCodes, res.StatusCode) && utils.StringArrayContains(c.Options.RetryOptions.Methods, method)) {
			return retry(res, err)
		}
	}

	if c.Options.FollowRedirect && res.Header.Get("location") != "" && utils.IntArrayContains(RedirectStatusCodes, res.StatusCode) {
		// we don't care about the response since we're redirecting
		res.Body.Close()

		if c.Options.RedirectOptions.Limit != 0 && len(c.Options.redirectUrls) >= c.Options.RedirectOptions.Limit {
			return nil, NewMaxRedirectsExceededError(len(c.Options.redirectUrls))
		}

		if c.Options.RedirectOptions.RewriteMethods || (res.StatusCode == 303 && c.Options.Method != "GET" && c.Options.Method != "HEAD") {
			c.Options.Method = "GET"
			c.CloseBody()
			c.Options.Headers.Del("content-length")
			c.Options.Headers.Del("content-type")
		}

		currentUrl := c.Options.FullUrl
		redirectUrl, err := utils.MergeUrl(currentUrl.String(), res.Header.Get("location"), true)
		if err != nil {
			return nil, err
		}

		// remove sensitive headers when redirecting to a different domain
		if redirectUrl.Hostname() != currentUrl.Hostname() || redirectUrl.Port() != currentUrl.Port() {
			c.Options.Headers.Del("host")
			c.Options.Headers.Del("cookie")
			c.Options.Headers.Del("authorization")
		}

		c.updateRequestCookies(res)
		c.Options.PrefixURL = ""
		c.Options.redirectUrls = append(c.Options.redirectUrls, redirectUrl)

		for _, hook := range c.Options.Hooks.BeforeRedirect {
			hook(c.Options, res)
		}

		return c.DoRequest(c.Options.Method, redirectUrl.String())
	}

	return res, nil
}

// Do is an alias of DoRequest.
func (c *Client) Do(method string, url string, options ...*Options) (*Response, error) {
	return c.DoRequest(method, url, options...)
}

func (c *Client) getTimeout(response *Response) (time.Duration, error) {
	retryAfter := strings.TrimSpace(response.Header.Get("retry-after"))

	// response header doesn't specify timeout, default to request timeout
	if retryAfter == "" || c.Options.RetryOptions.RetryAfter == false {
		return c.Options.Timeout, nil
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
func (c *Client) updateRequestCookies(response *Response) {
	cookies := c.getRequestCookies()
	if c.Options.CookieJar != nil && cookies != nil {
		// changed denotes whether or not a response cookie has a different value than a request cookie
		var changed bool
		for _, cookie := range response.Cookies() {
			if _, ok := cookies[cookie.Name]; ok {
				delete(cookies, cookie.Name)
				changed = true
			}
		}
		if changed {
			c.Options.Headers.Del("Cookie")
			var cks []string
			for _, cs := range cookies {
				for _, cookie := range cs {
					cks = append(cks, cookie.Name+"="+cookie.Value)
				}
			}
			c.Options.Headers.Set("Cookie", strings.Join(cks, "; "))
		}
	}
}

// getRequestCookies returns the cookies that were manually set in Options.Headers.
func (c *Client) getRequestCookies() map[string][]*http.Cookie {
	var cookies map[string][]*http.Cookie

	if c.Options.CookieJar != nil && c.Options.Headers.Get("cookie") != "" {
		cookies = make(map[string][]*http.Cookie)
		req := http.Request{Header: c.Options.Headers}
		for _, c := range req.Cookies() {
			cookies[c.Name] = append(cookies[c.Name], c)
		}
	}

	return cookies
}

// CloseBody clears the Body, Form and Json fields.
func (c *Client) CloseBody() {
	if c.Options.Body != nil {
		c.Options.Body.Close()
	}
	c.Options.Body = nil
	c.Options.Form = nil
	c.Options.Json = nil
}

// ParseBody parses Form or Json (in that order) into Content.
func (c *Client) ParseBody() error {
	if len(c.Options.Form) != 0 {
		encoded := c.Options.Form.EncodeWithOrder()
		c.Options.Body = io.NopCloser(strings.NewReader(encoded))
		return nil
	} else if j := c.Options.Json; len(j) != 0 {
		bytes, err := c.Options.MarshalJson(j)
		if err != nil {
			return err
		}
		c.Options.Body = io.NopCloser(bytesPkg.NewReader(bytes))
		return nil
	}
	return nil
}

func (c *Client) Get(url string, options ...*Options) (*Response, error) {
	return c.DoRequest("GET", url, options...)
}

func (c *Client) Post(url string, options ...*Options) (*Response, error) {
	return c.DoRequest("POST", url, options...)
}

func (c *Client) Update(url string, options ...*Options) (*Response, error) {
	return c.DoRequest("UPDATE", url, options...)
}

func (c *Client) Patch(url string, options ...*Options) (*Response, error) {
	return c.DoRequest("PATCH", url, options...)
}

func (c *Client) Delete(url string, options ...*Options) (*Response, error) {
	return c.DoRequest("DELETE", url, options...)
}

func (c *Client) Head(url string, options ...*Options) (*Response, error) {
	return c.DoRequest("HEAD", url, options...)
}
