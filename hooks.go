package gotcha

import "net/http"

type BeforeRequest func(*http.Request)

type BeforeRedirect func(*http.Request, *http.Response)

type AfterResponse func(response *http.Response, retry func())

type Hooks struct {
	BeforeRequest []BeforeRequest

	BeforeRedirect []BeforeRedirect

	AfterResponse []AfterResponse
}
