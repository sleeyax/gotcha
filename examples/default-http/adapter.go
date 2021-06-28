package main

import (
	"github.com/sleeyax/gotcha"
	"net/http"
)

// Adapter is a variable containing a customized gotcha.RequestAdapter.
var Adapter = &gotcha.RequestAdapter{
	RoundTripper: &http.Transport{
		ForceAttemptHTTP2:  true,
		DisableCompression: true,
	},
}
