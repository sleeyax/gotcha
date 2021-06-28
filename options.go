package gotcha

import (
	"encoding/json"
	"github.com/Sleeyax/urlValues"
	"github.com/imdario/mergo"
	"io"
	"net/http"
	"net/http/cookiejar"
	"net/url"
	"time"
)

var RedirectStatusCodes = []int{300, 301, 302, 303, 304, 307, 308}

type JSON map[string]interface{}

type UnmarshalJsonFunc = func(data []byte) (JSON, error)
type MarshalJsonFunc = func(json JSON) ([]byte, error)

type RedirectOptions struct {
	// Specifies if redirects should be rewritten as GET.
	//
	// If false, when sending a POST request and receiving a 302,
	// it will resend the body to the new location using the same HTTP method (POST in this case).
	//
	// Note that if a 303 is sent by the server in response to any request type (POST, DELETE, etc.),
	// gotcha will automatically request the resource pointed to in the location header via GET.
	// This is in accordance with the spec https://tools.ietf.org/html/rfc7231#section-6.4.4.
	RewriteMethods bool

	// Maximum amount of redirects to follow.
	// Follows an unlimited amount of redirects when set to 0.
	Limit int
}

type Options struct {
	// Adapter is an adapter that will be used by gotcha to make the actual request.
	// Implement your own Adapter or use the RequestAdapter to get started.
	Adapter Adapter

	// Request URI.
	// Can be relative or absolute.
	URI string

	// FullUrl is the URI that was computed form PrefixURL and URI.
	// You shouldn't need to modify this in most cases.
	FullUrl *url.URL

	// Proxy URL.
	// If this is an authenticated Proxy, make sure Username and Password are set.
	Proxy *url.URL

	// Retry on failure.
	Retry bool

	// Additional configuration options for Retry.
	RetryOptions *RetryOptions

	// Amount of retries that have been done so far.
	retries int

	// The HTTP method used to make the request.
	Method string

	// When specified, prefixUrl will be prepended to the url.
	// The prefix can be any valid URI, either relative or absolute.
	// A trailing slash / is optional - one will be added automatically.
	PrefixURL string

	// Request headers.
	Headers http.Header

	// Request Body.
	//
	// Body will be set in the following order,
	// whichever value is found to be of non-zero value first: Form -> Json -> Body.
	// Raw body content.
	Body io.ReadCloser

	// JSON data.
	Json JSON

	// Form data that will be converted to a query string.
	Form urlValues.Values

	// A function used to parse JSON responses.
	UnmarshalJson UnmarshalJsonFunc

	// A function used to stringify the body of JSON requests.
	MarshalJson MarshalJsonFunc

	// Can contain custom user data.
	// This can be  useful for storing authentication tokens for example.
	Context interface{}

	// CokieJar automatically stores & parses cookies.
	//
	// The CookieJar is used to insert relevant cookies into every
	// outbound Request and is updated with the cookie values
	// of every inbound Response. The CookieJar is also consulted for every
	// redirect that the Client follows.
	//
	// If CookieJar is nil, cookies are only sent if they are explicitly set on the Request.
	CookieJar http.CookieJar

	// Query string that will be added to the request URI.
	// This will override the query string in URI.
	SearchParams urlValues.Values

	// Duration to wait for the server to end the response before aborting the request.
	Timeout time.Duration

	// Defines if redirect responses should be followed automatically.
	FollowRedirect bool

	// Additional configuration options for FollowRedirect.
	RedirectOptions RedirectOptions

	// List of URls that have responded with a redirect so far.
	redirectUrls []*url.URL

	// Hooks allow modifications during the request lifecycle.
	Hooks Hooks
}

type RetryOptions struct {
	// Max number of times to retry.
	Limit int

	// Only retry when the request HTTP method equals one of these Methods.
	Methods []string

	// Only retry when the response HTTP status code equals one of these StatusCodes.
	StatusCodes []int

	// Only retry on error when the error message contains one of these ErrorCodes.
	ErrorCodes []string

	// Respect the response 'Retry-After' header, if set.
	//
	// If RetryAfter is false or the response headers don't contain this header,
	// it will default to the configured request Timeout.
	//
	// https://developer.mozilla.org/en-US/docs/Web/HTTP/Headers/Retry-After
	RetryAfter bool

	// CalculateTimeout is a function that computes the timeout to use between retries.
	// By default 'computedTimeout' will be used as timeout value.
	CalculateTimeout func(retries int, retryOptions *RetryOptions, computedTimeout time.Duration, error error) time.Duration
}

func NewDefaultOptions() *Options {
	jar, _ := cookiejar.New(&cookiejar.Options{})

	return &Options{
		URI:          "",
		Retry:        true,
		RetryOptions: NewDefaultRetryOptions(),
		Method:       "GET",
		PrefixURL:    "",
		Headers:      make(http.Header),
		Body:         nil,
		Json:         nil,
		Form:         nil,
		UnmarshalJson: func(data []byte) (JSON, error) {
			var result JSON
			if err := json.Unmarshal(data, &result); err != nil {
				return nil, err
			}
			return result, nil
		},
		MarshalJson: func(data JSON) ([]byte, error) {
			result, err := json.Marshal(data)
			if err != nil {
				return nil, err
			}
			return result, nil
		},
		Context:        nil,
		CookieJar:      jar,
		SearchParams:   nil,
		Timeout:        time.Second * 10,
		FollowRedirect: true,
		RedirectOptions: RedirectOptions{
			Limit:          0,
			RewriteMethods: true,
		},
		Hooks:   Hooks{},
		Adapter: &RequestAdapter{},
	}
}

func NewDefaultRetryOptions() *RetryOptions {
	return &RetryOptions{
		Limit:       2,
		Methods:     []string{http.MethodGet, http.MethodPut, http.MethodHead, http.MethodDelete, http.MethodOptions, http.MethodTrace},
		StatusCodes: []int{408, 413, 429, 500, 502, 503, 504, 521, 522, 524},
		ErrorCodes:  []string{"ETIMEDOUT", "ECONNRESET", "EADDRINUSE", "ECONNREFUSED", "EPIPE", "ENOTFOUND", "ENETUNREACH", "EAI_AGAIN"},
		RetryAfter:  true,
		CalculateTimeout: func(retries int, retryOptions *RetryOptions, computedTimeout time.Duration, error error) time.Duration {
			return computedTimeout
		},
	}
}

// Extend updates values from the current Options with values from the specified options.
func (o *Options) Extend(options *Options) (*Options, error) {
	dst := options

	// This check is necessary to fix weird 'panic: reflect: Field index out of range' error.
	// Mentioned error probably occurs because mergo can't decide how to merge Adapter structs that contain non-primitive fields
	// (i.e (nested) struct fields).
	//
	// The code below assures to just set the new adapter whenever it's not nil, skipping mergo's opeeration in the process.
	if options.Adapter != nil {
		o.Adapter = options.Adapter
	}

	if err := mergo.Merge(dst, o); err != nil {
		return nil, err
	}
	return dst, nil
}
