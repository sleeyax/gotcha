package gotcha

import (
	"encoding/json"
	"github.com/imdario/mergo"
	"io"
	"net/http"
	"net/http/cookiejar"
	"net/url"
)

type RedirectOptions struct {
	// Specifies if redirects should be rewritten as GET.
	//
	// If false, when sending a POST request and receiving a 302,
	// it will resend the body to the new location using the same HTTP method (POST in this case).
	RewriteMethods bool

	// If a 303 (See Other) status code is sent by the server in response to any request type (POST, DELETE, ...),
	// gotcha will automatically request the resource pointed to in the location header via GET
	// unless this field is set to false.
	HandleSeeOther bool

	// Maximum amount of redirects to follow.
	// Follows an unlimited amount of redirects when set to 0.
	MaxRedirects int
}

type Options struct {
	// Adapter is an adapter that will be used by gotcha to make the actual request.
	// Implement your own Adapter or use the RequestAdapter to get started.
	Adapter Adapter

	// Request URL.
	URL *url.URL

	// FullUrl is the URL that was computed form PrefixURL and URL.
	// You shouldn't need to modify this in most cases.
	FullUrl *url.URL

	// Retry on failure.
	Retry bool

	// The HTTP method used to make the request.
	Method string

	// When specified, prefixUrl will be prepended to the url.
	// The prefix can be any valid URL, either relative or absolute.
	// A trailing slash / is optional - one will be added automatically.
	PrefixURL string

	// Request headers.
	Headers http.Header

	// Request body.
	//
	// Body cannot be used with the Json or Form options.
	Body io.ReadCloser

	// JSON request Body.
	Json string

	// Form body request.
	// It will be converted to a query string.
	Form url.Values

	// Can contain custom user data.
	// It's useful for storing authentication tokens for example.
	Context interface{}

	// A function used to parse JSON responses.
	UnmarshalJson func(data []byte) (map[string]interface{}, error)

	// A function used to stringify the body of JSON requests.
	MarshalJson func(json map[string]interface{}) ([]byte, error)

	// Automatically store & parse cookies.
	CookieJar http.CookieJar

	// Query string that will be added to the request URL.
	// This will override the query string in URL.
	SearchParams url.Values

	// Milliseconds to wait for the server to end the response before aborting the request.
	Timeout int

	// Defines if redirect responses should be followed automatically.
	FollowRedirect bool

	// Additional configuration options for FollowRedirect.
	RedirectOptions RedirectOptions

	// Middleware functions.
	Hooks Hooks
}

func NewDefaultOptions() *Options {
	jar, _ := cookiejar.New(&cookiejar.Options{})

	return &Options{
		URL:       nil,
		Retry:     true,
		Method:    "GET",
		PrefixURL: "",
		Headers:   make(http.Header),
		Body:      nil,
		Json:      "",
		Form:      nil,
		Context:   nil,
		UnmarshalJson: func(data []byte) (map[string]interface{}, error) {
			var result map[string]interface{}
			if err := json.Unmarshal(data, &result); err != nil {
				return nil, err
			}
			return result, nil
		},
		MarshalJson: func(data map[string]interface{}) ([]byte, error) {
			result, err := json.Marshal(data)
			if err != nil {
				return nil, err
			}
			return result, nil
		},
		CookieJar:      jar,
		SearchParams:   nil,
		Timeout:        10000,
		FollowRedirect: false,
		RedirectOptions: RedirectOptions{
			HandleSeeOther: true,
			MaxRedirects:   0,
			RewriteMethods: false,
		},
		Hooks:   Hooks{},
		Adapter: &RequestAdapter{},
	}
}

// Extend updates values from the current Options with values from the specified options.
func (o *Options) Extend(options *Options) (*Options, error) {
	dst := options
	if err := mergo.Merge(dst, o); err != nil {
		return nil, err
	}
	return dst, nil
}
