package utils

import (
	"github.com/sleeyax/gotcha/internal/tests"
	"testing"
)

func TestMergeUrl(t *testing.T) {
	u1 := "https://example.com/a/b"
	u2 := "https://domain.example.com/b/c"

	url, err := MergeUrl(u1, u2, true)
	if err != nil {
		t.Fatal(err)
	}

	if u := url.String(); u != u2 {
		t.Fatalf(tests.MismatchFormat, "url", u2, u)
	}

	u2 = "/foo/bar"

	url, err = MergeUrl(u1, u2, true)
	if err != nil {
		t.Fatal(err)
	}

	if u := url.String(); u != "https://example.com/foo/bar" {
		t.Fatalf(tests.MismatchFormat, "url", "https://example.com/foo/bar", u)
	}

	u1 = ""
	u2 = "https://example.com"

	url, err = MergeUrl(u1, u2, true)
	if err != nil {
		t.Fatal(err)
	}

	if u := url.String(); u != u2 {
		t.Fatalf(tests.MismatchFormat, "url", u2, u)
	}
}
