package gotcha

import (
	"fmt"
	"github.com/sleeyax/gotcha/internal/tests"
	"net/http"
	"testing"
)

type MockAdapter struct {
	OnCalledDoRequest func()
}

func (ma *MockAdapter) DoRequest(options *Options) (*http.Response, error) {
	ma.OnCalledDoRequest()
	return &http.Response{StatusCode: 200}, nil
}

func TestNew(t *testing.T) {
	var requested bool

	res, err := New("https://example.com", "GET", &Options{
		Adapter: &MockAdapter{OnCalledDoRequest: func() {
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
