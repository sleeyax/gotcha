package gotcha

import (
	"fmt"
	"github.com/sleeyax/gotcha/internal/tests"
	"testing"
)

func TestDoRequest(t *testing.T) {
	var requested bool

	res, err := DoRequest("https://example.com", "GET", &Options{
		Adapter: &mockAdapter{OnCalledDoRequest: func() {
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

func ExamplePost() {
	// TODO: improve example
	res, err := Post("https://httpbin.org/get", NewDefaultOptions())
	if err != nil {
		fmt.Println(err)
	}
	fmt.Println(res.StatusCode)
}
