package gotcha

import (
	"github.com/sleeyax/gotcha/internal/tests"
	"io"
	"net/http"
	"strings"
	"testing"
)

func TestDoRequest(t *testing.T) {
	var requested bool

	res, err := DoRequest("https://example.com", "GET", &Options{
		Adapter: &mockAdapter{OnCalledDoRequest: func(_ *Options) *http.Response {
			requested = true
			return &http.Response{StatusCode: 200}
		}},
	})
	if err != nil {
		t.Fatal(err)
	}

	if requested != true {
		t.Fatalf("adapter failed to execute the request")
	}

	if s := res.StatusCode; s != 200 {
		t.Fatalf(tests.MismatchFormat, "status code", 200, s)
	}
}

func TestPost(t *testing.T) {
	createBody := func(str string) io.ReadCloser {
		return io.NopCloser(strings.NewReader(str))
	}
	expectedBody := "bar"

	// test that the last option body we provided is used when performing the request
	_, err := Post(
		"https://example.com",
		&Options{Body: createBody("foo"), Adapter: &mockAdapter{OnCalledDoRequest: func(o *Options) *http.Response {
			body, _ := io.ReadAll(o.Body)
			if b := string(body); b != expectedBody {
				t.Errorf(tests.MismatchFormat, "body", expectedBody, b)
			}
			return &http.Response{StatusCode: 200}
		}}},
		&Options{Body: createBody(expectedBody)},
	)
	if err != nil {
		t.Fatal(err)
	}
}

func TestGet(t *testing.T) {
	_, err := Get("https://example.com", &Options{Adapter: &mockAdapter{OnCalledDoRequest: func(options *Options) *http.Response {
		if options.Method != "GET" {
			t.Fatalf(tests.MismatchFormat, "method", "GET", options.Method)
		}
		return &http.Response{StatusCode: 200}
	}}})
	if err != nil {
		t.Fatal(err)
	}
}
