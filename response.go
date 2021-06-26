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
