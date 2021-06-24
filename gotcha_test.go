package gotcha

import (
	"github.com/sleeyax/gotcha/internal/tests"
	"io"
	"strings"
	"testing"
)

func TestDoRequest(t *testing.T) {
	var requested bool

	res, err := DoRequest("https://example.com", "GET", &Options{
		Adapter: &mockAdapter{OnCalledDoRequest: func(_ *Options) {
			requested = true
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
		&Options{Body: createBody("foo"), Adapter: &mockAdapter{OnCalledDoRequest: func(o *Options) {
			body, _ := io.ReadAll(o.Body)
			if b := string(body); b != expectedBody {
				t.Errorf(tests.MismatchFormat, "body", expectedBody, b)
			}
		}}},
		&Options{Body: createBody(expectedBody)},
	)

	if err != nil {
		t.Fatal(err)
	}
}
