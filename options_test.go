package gotcha

import (
	"github.com/sleeyax/gotcha/internal/tests"
	"io"
	"net/http"
	"net/url"
	"strings"
	"testing"
)

func TestOptions_Merge(t *testing.T) {
	var beforeRequestCalled bool

	left := NewDefaultOptions()
	right := &Options{
		Url:    &url.URL{Host: "example.com"},
		Retry:  true,
		Method: "POST",
		Headers: http.Header{
			"Foo": {"Bar"},
		},
		Body:          io.NopCloser(strings.NewReader("hello world")),
		UnmarshalJson: nil,
		MarshalJson:   nil,
		SearchParams: url.Values{
			"abc": {"def"},
		},
		Timeout:         1000,
		FollowRedirect:  false,
		RedirectOptions: RedirectOptions{},
		Hooks: Hooks{
			BeforeRequest: []BeforeRequest{
				func(_ *http.Request) {
					beforeRequestCalled = true
				},
			},
		},
	}

	options, err := left.Extend(right)
	if err != nil {
		t.Fatal(err)
	}

	// test merged options fields
	if u := options.Url; u != right.Url {
		t.Errorf(tests.MismatchFormat, "url", right.Url, u)
	}

	if r := options.Retry; r != right.Retry {
		t.Errorf(tests.MismatchFormat, "retry", right.Retry, r)
	}

	if m := options.Method; m != right.Method {
		t.Errorf(tests.MismatchFormat, "method", right.Method, m)
	}

	if h := options.Headers.Get("Foo"); h != "Bar" {
		t.Errorf(tests.MismatchFormat, "headers", "Bar", h)
	}

	optionsBody, err := io.ReadAll(options.Body)
	if err != nil {
		t.Error(err)
	}
	if b := string(optionsBody); b != "hello world" {
		t.Errorf(tests.MismatchFormat, "body", "hello world", b)
	}

	if options.UnmarshalJson == nil || options.MarshalJson == nil {
		t.Errorf("Any of the json marshal functions shouldn't be nil.")
	}

	if sp := options.SearchParams.Encode(); sp != "abc=def" {
		t.Errorf(tests.MismatchFormat, "search parameters", "abc=def", sp)
	}

	if to := options.Timeout; to != right.Timeout {
		t.Errorf(tests.MismatchFormat, "timeout", right.Timeout, to)
	}

	if fr := options.FollowRedirect; fr != right.FollowRedirect {
		t.Errorf(tests.MismatchFormat, "follow redirect", right.FollowRedirect, fr)
	}

	if ro := options.RedirectOptions; ro != right.RedirectOptions {
		t.Errorf(tests.MismatchFormat, "redirect options", right.RedirectOptions, ro)
	}

	if len(options.Hooks.BeforeRequest) == 0 {
		t.Errorf("At least 1 BeforeRequest hook should be set.")
	}
	options.Hooks.BeforeRequest[0](&http.Request{})
	if !beforeRequestCalled {
		t.Errorf(tests.MismatchFormat, "hook result 'beforeRequestCalled'", true, beforeRequestCalled)
	}
}
