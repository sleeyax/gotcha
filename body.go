package gotcha

import (
	bytes2 "bytes"
	"github.com/Sleeyax/urlValues"
	"io"
	"strings"
)

type Body struct {
	// Raw body content.
	Content io.ReadCloser

	// JSON data.
	Json map[string]interface{}

	// Form data that will be converted to a query string.
	Form urlValues.Values

	// A function used to parse JSON responses.
	UnmarshalJson func(data []byte) (map[string]interface{}, error)

	// A function used to stringify the body of JSON requests.
	MarshalJson func(json map[string]interface{}) ([]byte, error)
}

// Close clears the Content, Form and Json fields.
func (b *Body) Close() {
	if b.Content != nil {
		b.Content.Close()
	}
	b.Content = nil
	b.Form = nil
	b.Json = nil
}

// Parse parses Form or Json (in that order) into Content.
func (b *Body) Parse() error {
	if len(b.Form) != 0 {
		encoded := b.Form.EncodeWithOrder()
		b.Content = io.NopCloser(strings.NewReader(encoded))
		return nil
	} else if j := b.Json; len(j) != 0 {
		bytes, err := b.MarshalJson(j)
		if err != nil {
			return err
		}
		b.Content = io.NopCloser(bytes2.NewReader(bytes))
		return nil
	}
	return nil
}
