package gotcha

import (
	"errors"
	"github.com/sleeyax/gotcha/internal/utils"
	"net/http"
	urlPkg "net/url"
	"strconv"
	"strings"
	"time"
)

var RequestFailedError = errors.New("request failed and max retries exceeded")

type Client struct {
	options *Options
}

// NewClient creates a new HTTP client based on default Options which can be extended.
func NewClient(options *Options) (*Client, error) {
	client := &Client{NewDefaultOptions()}
	return client.Extend(options)
}

// Extend returns a new Client with merged Options.
func (c *Client) Extend(options *Options) (*Client, error) {
	opts, err := c.options.Extend(options)
	if err != nil {
		return nil, err
	}
	return &Client{opts}, nil
}

func (c *Client) DoRequest(url string, method string) (*http.Response, error) {
	u, err := c.getFullUrl(url)
	if err != nil {
		return nil, err
	}

	c.options.FullUrl = u
	c.options.Method = method

	if sp := c.options.SearchParams; len(sp) != 0 {
		c.options.FullUrl.RawQuery = sp.EncodeWithOrder()
	}

	// TODO: body, json, form

	retry := func(res *http.Response, err error) (*http.Response, error) {
		timeout, e := c.getTimeout(res)
		if e != nil {
			return nil, e
		}
		timeout = c.options.RetryOptions.CalculateTimeout(c.options.retries, c.options.RetryOptions, timeout, err)
		time.Sleep(timeout)
		c.options.retries++
		return c.DoRequest(url, method)
	}

	res, err := c.options.Adapter.DoRequest(c.options)
	if err != nil {
		if c.options.Retry && c.options.retries < c.options.RetryOptions.Limit && utils.StringArrayContains(c.options.RetryOptions.ErrorCodes, err.Error()) {
			return retry(res, err)
		}
		return nil, err
	}

	if c.options.Retry && utils.IntArrayContains(c.options.RetryOptions.StatusCodes, res.StatusCode) && utils.StringArrayContains(c.options.RetryOptions.Methods, method) {
		if c.options.retries >= c.options.RetryOptions.Limit {
			return res, RequestFailedError
		}
		return retry(res, nil)
	}

	// TODO: cookiejar

	return res, nil
}

func (c *Client) getTimeout(response *http.Response) (time.Duration, error) {
	retryAfter := strings.TrimSpace(response.Header.Get("retry-after"))

	// response header doesn't specify timeout, default to request timeout
	if retryAfter == "" || c.options.RetryOptions.RetryAfter == false {
		return c.options.Timeout, nil
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

// getFullUrl computes the actual request url by combining prefixUrl and url.
func (c *Client) getFullUrl(url string) (*urlPkg.URL, error) {
	if c.options.PrefixURL == "" {
		return urlPkg.Parse(url)
	}

	u, err := urlPkg.Parse(c.options.PrefixURL)
	if err != nil {
		return nil, err
	}

	return u.Parse(url)
}

func (c *Client) Get(url string) (*http.Response, error) {
	return c.DoRequest(url, "GET")
}

func (c *Client) Post(url string) (*http.Response, error) {
	return c.DoRequest(url, "POST")
}

func (c *Client) Update(url string) (*http.Response, error) {
	return c.DoRequest(url, "UPDATE")
}

func (c *Client) Patch(url string) (*http.Response, error) {
	return c.DoRequest(url, "PATCH")
}

func (c *Client) Delete(url string) (*http.Response, error) {
	return c.DoRequest(url, "DELETE")
}

func (c *Client) Head(url string) (*http.Response, error) {
	return c.DoRequest(url, "HEAD")
}
