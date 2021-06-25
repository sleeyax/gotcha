package gotcha

import (
	"github.com/Sleeyax/urlValues"
	"github.com/sleeyax/gotcha/internal/tests"
	"io"
	"strings"
	"testing"
)

func TestBody_Parse(t *testing.T) {
	form := urlValues.Values{}
	form.Add("x", "y")

	body := Body{
		Content: io.NopCloser(strings.NewReader("foo is at the bar")),
		Json:    map[string]interface{}{"a": []string{"b"}},
		Form:    form,
	}

	body.Parse()

	contentBytes, err := io.ReadAll(body.Content)
	if err != nil {
		t.Fatal(err)
	}
	content := string(contentBytes)

	if content != "x=y" {
		t.Fatalf(tests.MismatchFormat, "body content", "x=y", content)
	}
}
