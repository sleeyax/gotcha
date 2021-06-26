package gotcha

import (
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
