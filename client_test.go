package gotcha

import (
	"encoding/json"
	"github.com/Sleeyax/urlValues"
	"github.com/sleeyax/gotcha/internal/tests"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"
)

func TestClient_Extend(t *testing.T) {
	client, err := NewClient(&Options{
		PrefixURL: "https://before.example.com",
	})
	if err != nil {
		t.Fatal(err)
	}

	prefixUrl := "https://after.example.com"

	client2, err := client.Extend(&Options{
		PrefixURL: prefixUrl,
	})
	if err != nil {
		t.Fatal(err)
	}

	if client2.options.PrefixURL != prefixUrl {
		t.FailNow()
	}
}

func TestClient_DoRequest_RetryAfter(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Add("retry-after", "3")
		w.WriteHeader(500)
	}))
	defer ts.Close()

	retriesLeft := 2

	client, err := NewClient(&Options{
		Retry: true,
		RetryOptions: &RetryOptions{
			Limit:       retriesLeft,
			Methods:     []string{"GET"},
			StatusCodes: []int{500},
			ErrorCodes:  []string{},
			RetryAfter:  true,
			CalculateTimeout: func(retries int, retryOptions *RetryOptions, computedTimeout time.Duration, error error) time.Duration {
				if s := computedTimeout.Seconds(); s != 3 {
					t.Fatalf(tests.MismatchFormat, "timeout", 3, s)
				}
				retriesLeft--
				return 0
			},
		},
	})
	if err != nil {
		t.Fatal(err)
	}

	res, err := client.DoRequest(ts.URL, "GET")
	if err == nil {
		t.Fatalf("request should have failed, but got status code %d", res.StatusCode)
	}
	if err != RequestFailedError {
		t.Fatal(err)
	}

	if retriesLeft != 0 {
		t.Fatalf("there should be 0 retries left, but there are %d", retriesLeft)
	}
}

func TestClient_DoRequest_Body(t *testing.T) {
	url := "https://example.com"
	var testType string
	var wantedBody string

	client, err := NewClient(&Options{
		Adapter: &mockAdapter{OnCalledDoRequest: func(options *Options) *http.Response {
			bodyBytes, err := io.ReadAll(options.Body.Content)
			if err != nil {
				t.Fatalf("failed to read body while testing %s", testType)
			}
			body := string(bodyBytes)
			if body != wantedBody {
				t.Fatalf(tests.MismatchFormat, testType, wantedBody, body)
			}
			return &http.Response{StatusCode: 200}
		}},
	})
	if err != nil {
		t.Fatal(err)
	}

	testType = "raw body"
	wantedBody = "hello world!"
	client.options.Body.Content = io.NopCloser(strings.NewReader(wantedBody))
	client.Post(url)

	testType = "form"
	wantedBody = "foo=bar&abc=def"
	client.options.Body.Form = urlValues.Values{
		"foo":              {"bar"},
		"abc":              {"def"},
		urlValues.OrderKey: []string{"foo", "abc"},
	}
	client.Post(url)

	testType = "json"
	wantedBody = `{"a":"b","c":["d","e","f"],"g":{"h":"i"}}`
	var result map[string]interface{}
	json.Unmarshal([]byte(wantedBody), &result)
	client.options.Body.Json = result
	client.Post(url)
}
