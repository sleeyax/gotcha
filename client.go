package gotcha

import (
	"net/http"
	urlPkg "net/url"
)

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
	u, err := urlPkg.Parse(url)
	if err != nil {
		return nil, err
	}

	c.options.Url = u
	c.options.Method = method

	return c.options.Adapter.DoRequest(c.options)
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
