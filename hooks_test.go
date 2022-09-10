package gotcha

import (
	"bytes"
	"github.com/sleeyax/gotcha/internal/tests"
	"io"
	"net/http"
	"strings"
	"testing"
)

func createClient(t *testing.T, hooks Hooks) *Client {
	client, err := NewClient(&Options{
		Hooks: hooks,
		Adapter: &mockAdapter{OnCalledDoRequest: func(options *Options) *Response {
			return NewResponse(&http.Response{
				Request: &http.Request{
					Method: options.Method,
					URL:    options.FullUrl,
					Body:   options.Body,
					Header: options.Headers,
				},
			})
		}},
	})
	if err != nil {
		t.Fatal(err)
	}

	return client
}

func TestHooks_Init(t *testing.T) {
	var hooked bool

	client := createClient(t, Hooks{
		Init: []InitHook{
			func(options *Options) {
				hooked = true
			},
		},
	})
	client.Get("https://example.com")

	if !hooked {
		t.FailNow()
	}
}

func TestHooks_BeforeRequest(t *testing.T) {
	var hooked bool

	client := createClient(t, Hooks{
		BeforeRequest: []BeforeRequestHook{
			func(options *Options) {
				hooked = true
				options.Method = http.MethodPost
			},
		},
	})
	res, err := client.Get("https://example.com")
	if err != nil {
		t.FailNow()
	}

	if !hooked {
		t.FailNow()
	}

	if m := res.Request.Method; m != http.MethodPost {
		t.Fatalf(tests.MismatchFormat, "method", http.MethodPost, m)
	}
}

func TestHooks_BeforeRedirect(t *testing.T) {
	var hooked bool
	prefixUrl := "https://redirected.example.com"

	client, err := NewClient(&Options{
		Hooks: Hooks{
			BeforeRedirect: []BeforeRedirectHook{
				func(options *Options, response *Response) {
					hooked = true
					options.PrefixURL = prefixUrl
				},
			},
		},
		Adapter: &mockAdapter{OnCalledDoRequest: func(options *Options) *Response {
			header := http.Header{}
			header.Add("location", "/home")
			return NewResponse(&http.Response{StatusCode: 302, Header: header, Body: io.NopCloser(bytes.NewReader([]byte{}))})
		}},
		FollowRedirect: true,
		RedirectOptions: RedirectOptions{
			RewriteMethods: false,
			Limit:          1,
		},
	})
	if err != nil {
		t.Fatal(err)
	}
	client.Get("https://example.com")

	if !hooked {
		t.FailNow()
	}

	if u := client.Options.FullUrl.String(); u != prefixUrl+"/home" {
		t.Fatalf(tests.MismatchFormat, "full url", prefixUrl+"/home", u)
	}
}

func TestHooks_BeforeRetry_And_AfterResponse(t *testing.T) {
	var hookedBeforeRetry bool
	var hookedAfterResponse bool

	client := createClient(t, Hooks{
		BeforeRetry: []BeforeRetryHook{
			func(options *Options, error error, retryCount int) {
				hookedBeforeRetry = true
			},
		},
		AfterResponse: []AfterResponseHook{
			func(response *Response, retry RetryFunc) (*Response, error) {
				hookedAfterResponse = true
				response.Body = io.NopCloser(strings.NewReader("hijacked"))
				if !hookedBeforeRetry {
					return retry(&Options{})
				}
				return response, nil
			},
		},
	})
	res, _ := client.Get("https://example.com")

	if !hookedBeforeRetry || !hookedAfterResponse {
		t.FailNow()
	}

	bodyBytes, err := io.ReadAll(res.Body)
	if err != nil {
		t.Fatal(err)
	}
	body := string(bodyBytes)

	if body != "hijacked" {
		t.Fatalf(tests.MismatchFormat, "body", "hijacked", body)
	}
}
