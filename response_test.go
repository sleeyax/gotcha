package gotcha

import (
	"bytes"
	"fmt"
	"github.com/sleeyax/gotcha/internal/tests"
	"io"
	"net/http"
	"strings"
	"testing"
)

func TestResponse_Json(t *testing.T) {
	res := NewResponse(&http.Response{
		StatusCode: 200,
		Body:       io.NopCloser(strings.NewReader(`{"foo": "bar"}`)),
	})
	res.UnmarshalJsonFunc = NewDefaultOptions().UnmarshalJson

	json, err := res.Json()
	if err != nil {
		t.Fatal(err)
	}

	if value, ok := json["foo"]; !ok || value != "bar" {
		t.Fatalf(tests.MismatchFormat, "json value", "bar", fmt.Sprintf("%s (ok: %v)", value, ok))
	}
}

func TestResponse_Raw(t *testing.T) {
	body := []byte{1, 2, 3}

	res := NewResponse(&http.Response{
		StatusCode: 200,
		Body:       io.NopCloser(bytes.NewReader(body)),
	})

	b, err := res.Raw()
	if err != nil {
		t.Fatal(err)
	}

	if bytes.Compare(b, body) != 0 {
		t.FailNow()
	}
}

func TestResponse_Text(t *testing.T) {
	body := "hello world"

	res := NewResponse(&http.Response{
		StatusCode: 200,
		Body:       io.NopCloser(strings.NewReader(body)),
	})

	text, err := res.Text()
	if err != nil {
		t.Fatal(err)
	}

	if text != body {
		t.FailNow()
	}
}
