package gotcha

import (
	bytes2 "bytes"
	"encoding/json"
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
// The resulting Content value will also be returned as a string by this function.
func (b *Body) Parse() (string, error) {
	if len(b.Form) != 0 {
		encoded := b.Form.EncodeWithOrder()
		b.Content = io.NopCloser(strings.NewReader(encoded))
		return encoded, nil
	} else if j := b.Json; len(j) != 0 {
		bytes, err := json.Marshal(j)
		if err != nil {
			return "", err
		}
		b.Content = io.NopCloser(bytes2.NewReader(bytes))
		return string(bytes), nil
	}
	return "", nil
}
