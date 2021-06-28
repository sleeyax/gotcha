package main

import (
	"fmt"
	"github.com/sleeyax/gotcha"
	"log"
	"net/http"
)

// CustomRequestAdapter is a variable containing a customized gotcha.RequestAdapter.
var CustomRequestAdapter = &gotcha.RequestAdapter{
	RoundTripper: &http.Transport{
		ForceAttemptHTTP2:  true,
		DisableCompression: true,
	},
}

func main() {
	res, err := gotcha.Get("https://sleeyax.dev", &gotcha.Options{
		Adapter: CustomRequestAdapter,
	})
	if err != nil {
		log.Fatal(err)
	}
	fmt.Println(res.Text())
	// Output: <!DOCTYPE html>...
}
