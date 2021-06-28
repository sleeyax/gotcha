package gotcha

import (
	"io"
	"net/http"
)

type Response struct {
	*http.Response
	UnmarshalJsonFunc
}

// Json parses the Response Body as JSON.
func (r *Response) Json() (JSON, error) {
	bb, err := io.ReadAll(r.Body)
	if err != nil {
		return nil, err
	}
	return r.UnmarshalJsonFunc(bb)
}

// Raw reads the Response Body as a byte array.
func (r *Response) Raw() ([]byte, error) {
	bb, err := io.ReadAll(r.Body)
	if err != nil {
		return nil, err
	}
	return bb, nil
}

// Text reads the Response Body as a string.
func (r *Response) Text() (string, error) {
	raw, err := r.Raw()
	if err != nil {
		return "", err
	}
	return string(raw), nil
}

func NewResponse(response *http.Response) *Response {
	return &Response{
		Response:          response,
		UnmarshalJsonFunc: nil,
	}
}

// Close closes the response body.
func (r *Response) Close() error {
	return r.Body.Close()
}
