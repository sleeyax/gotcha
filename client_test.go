package gotcha

import (
	"encoding/json"
	"fmt"
	"github.com/Sleeyax/urlValues"
	"github.com/sleeyax/gotcha/internal/tests"
	"io"
	"net/http"
	"net/http/cookiejar"
	"net/http/httptest"
	urlPkg "net/url"
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

	if client2.Options.PrefixURL != prefixUrl {
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
			Methods:     []string{http.MethodGet},
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

	res, err := client.DoRequest(http.MethodGet, ts.URL)
	if err == nil {
		t.Fatalf("request should have failed, but got status code %d", res.StatusCode)
	}
	if _, ok := err.(*MaxRetriesExceededError); !ok {
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
		Adapter: &mockAdapter{OnCalledDoRequest: func(options *Options) *Response {
			bodyBytes, err := io.ReadAll(options.Body)
			if err != nil {
				t.Fatalf("failed to read body while testing %s", testType)
			}
			body := string(bodyBytes)
			if body != wantedBody {
				t.Fatalf(tests.MismatchFormat, testType, wantedBody, body)
			}
			return NewResponse(&http.Response{StatusCode: 200})
		}},
	})
	if err != nil {
		t.Fatal(err)
	}

	testType = "raw body"
	wantedBody = "hello world!"
	client.Options.Body = io.NopCloser(strings.NewReader(wantedBody))
	client.Post(url)

	testType = "form"
	wantedBody = "foo=bar&abc=def"
	client.Options.Form = urlValues.Values{
		"foo":              {"bar"},
		"abc":              {"def"},
		urlValues.OrderKey: []string{"foo", "abc"},
	}
	client.Post(url)

	testType = "json"
	wantedBody = `{"a":"b","c":["d","e","f"],"g":{"h":"i"}}`
	var result JSON
	json.Unmarshal([]byte(wantedBody), &result)
	client.Options.Json = result
	client.Post(url)
}

func TestClient_DoRequest_Cookies(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch r.RequestURI {
		case "/":
			http.Redirect(w, r, "/cookie", 302)
			break
		case "/nocookie":
			w.WriteHeader(200)
			break
		case "/cookie":
			http.SetCookie(w, &http.Cookie{Name: "foo", Value: "bar"})
			w.WriteHeader(200)
			break
		case "/cookieredirect":
			http.SetCookie(w, &http.Cookie{Name: "abc", Value: "def"})
			http.Redirect(w, r, "/nocookie", 302)
			break
		}
	}))
	defer ts.Close()
	tsUrl, _ := urlPkg.Parse(ts.URL)

	jar, _ := cookiejar.New(&cookiejar.Options{})

	client, err := NewClient(&Options{
		CookieJar:      jar,
		FollowRedirect: true,
		RedirectOptions: RedirectOptions{
			RewriteMethods: false,
			Limit:          1,
		},
	})
	if err != nil {
		t.Fatal(err)
	}

	_, err = client.Get(ts.URL)
	if err != nil {
		t.Fatal(err)
	}

	// test if cookie was set after redirect has been followed
	cookies := jar.Cookies(tsUrl)
	if len(cookies) == 0 {
		t.Fatalf("0 cookies were set")
	}
	if firstCookie := cookies[0]; firstCookie.Name != "foo" && firstCookie.Value != "bar" {
		t.Fatalf(tests.MismatchFormat, "cookie", "foo=bar", firstCookie.Name+"="+firstCookie.Value)
	}

	// test override cookie
	jar, _ = cookiejar.New(&cookiejar.Options{})
	headers := http.Header{}
	headers.Add("cookie", "abc=xyz")
	client, _ = client.Extend(&Options{
		CookieJar: jar,
		Headers:   headers,
		Adapter: &mockAdapter{OnCalledDoRequest: func(options *Options) *Response {
			requestAdapter := RequestAdapter{}
			res, err := requestAdapter.DoRequest(options)
			if err != nil {
				t.Fatal(err)
			}
			if cookie := options.Headers.Get("cookie"); res.StatusCode == 200 && cookie != "abc=def" {
				t.Fatalf(tests.MismatchFormat, "cookie override", "abc=def", cookie)
			}
			return res
		}},
	})
	client.Get(ts.URL + "/cookieredirect")
}

func TestClient_DoRequest_Redirect(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch r.RequestURI {
		case "/":
			http.Redirect(w, r, "/home", 302)
			break
		case "/home":
			w.WriteHeader(200)
			break
		case "/loop":
			http.Redirect(w, r, "/loop", 302)
		}
	}))
	defer ts.Close()

	jar, _ := cookiejar.New(&cookiejar.Options{})

	client, err := NewClient(&Options{
		CookieJar:      jar,
		FollowRedirect: true,
		RedirectOptions: RedirectOptions{
			RewriteMethods: true,
			Limit:          3,
		},
	})
	if err != nil {
		t.Fatal(err)
	}

	// test rewrite methods
	client.Options.Body = io.NopCloser(strings.NewReader("hello world!"))
	res, err := client.Post(ts.URL + "/")
	if err != nil {
		t.Fatal(err)
	}
	if m := res.Request.Method; m != http.MethodGet {
		t.Fatalf(tests.MismatchFormat, "method", http.MethodGet, m)
	}

	// test retry limit
	_, err = client.Get(ts.URL + "/loop")
	if err == nil {
		t.Fatalf("expected MaxRedirectsExceededError to be non-nil")
	}
	if _, ok := err.(*MaxRedirectsExceededError); !ok {
		t.Fatal(err)
	}
}

func ExampleNewClient() {
	client, err := NewClient(&Options{
		PrefixURL: "https://httpbin.org/",
		Headers: http.Header{
			"user-agent": {"gotcha"},
		},
	})
	if err != nil {
		fmt.Sprintln("error:", err)
	}

	res, err := client.Do(http.MethodGet, "https://httpbin.org/get")
	if err != nil {
		fmt.Sprintln("error:", err)
	}

	json, _ := res.Json()
	headers := json["headers"].(map[string]interface{})

	fmt.Println(res.StatusCode)
	fmt.Println(headers["User-Agent"])
}
