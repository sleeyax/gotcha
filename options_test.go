package gotcha

import (
	urlValues "github.com/Sleeyax/urlValues"
	"github.com/sleeyax/gotcha/internal/tests"
	"io"
	"net/http"
	"strings"
	"testing"
)

func TestOptions_Merge(t *testing.T) {
	var beforeRequestCalled bool

	left := NewDefaultOptions()
	right := &Options{
		URI:    "example.com",
		Retry:  true,
		Method: http.MethodPost,
		Headers: http.Header{
			"Foo": {"Bar"},
		},
		Body: io.NopCloser(strings.NewReader("hello world")),
		SearchParams: urlValues.Values{
			"xyz":              {"123"},
			"abc":              {"def"},
			urlValues.OrderKey: {"xyz", "abc"},
		},
		Timeout:         1000,
		FollowRedirect:  false,
		RedirectOptions: RedirectOptions{},
		Hooks: Hooks{
			BeforeRequest: []BeforeRequestHook{
				func(options *Options) {
					beforeRequestCalled = true
				},
			},
		},
	}

	options, err := left.Extend(right)
	if err != nil {
		t.Fatal(err)
	}

	// test merged Options fields
	if u := options.URI; u != right.URI {
		t.Errorf(tests.MismatchFormat, "url", right.URI, u)
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

	if sp := options.SearchParams.EncodeWithOrder(); sp != "xyz=123&abc=def" {
		t.Errorf(tests.MismatchFormat, "search parameters", "abc=def", sp)
	}

	if to := options.Timeout; to != right.Timeout {
		t.Errorf(tests.MismatchFormat, "timeout", right.Timeout, to)
	}

	if fr := options.FollowRedirect; fr != right.FollowRedirect {
		t.Errorf(tests.MismatchFormat, "follow redirect", right.FollowRedirect, fr)
	}

	if ro := options.RedirectOptions; ro != right.RedirectOptions {
		t.Errorf(tests.MismatchFormat, "redirect Options", right.RedirectOptions, ro)
	}

	if len(options.Hooks.BeforeRequest) == 0 {
		t.Errorf("At least 1 BeforeRequestHook hook should be set.")
	}
	options.Hooks.BeforeRequest[0](&Options{})
	if !beforeRequestCalled {
		t.Errorf(tests.MismatchFormat, "hook result 'beforeRequestCalled'", true, beforeRequestCalled)
	}
}

func TestOptions_Merge_Bool(t *testing.T) {
	testCases := [2]bool{true, false}

	for _, x := range testCases {
		parent := &Options{Retry: x}

		for _, y := range testCases {
			child := &Options{Retry: y}

			merged, err := parent.Extend(child)

			if err != nil {
				t.Error(err)
			}

			if merged.Retry != child.Retry {
				t.Errorf(tests.MismatchFormat, "retry", child.Retry, merged.Retry)
			}
		}
	}
}
