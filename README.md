<h1 align="center">
  <img width="150" src="docs/assets/logo.png" />
  <p>gotcha</p>
</h1>

[![Go Reference](https://pkg.go.dev/badge/github.com/sleeyax/gotcha.svg)](https://pkg.go.dev/github.com/sleeyax/gotcha)

Gotcha is an alternative to Go's [http client](https://golang.org/src/net/http/client.go), 
with an API inspired by [got](https://github.com/sindresorhus/got).
It can interface with other HTTP packages through an adapter.

Aadapter implementations for [fhttp](https://github.com/useflyent/fhttp), [cclient](https://github.com/x04/cclient) & [fasthttp](https://github.com/valyala/fasthttp) can be found in the [adapters](adapters) directory.

## Usage
### Top-Level API
Gotcha exposes a top-level API to make quick and simple requests:
```go
package main

import (
	"fmt"
	"github.com/sleeyax/gotcha"
	"log"
)

func main() {
	res, err := gotcha.Get("https://sleeyax.dev")
	if err != nil {
		log.Fatal(err)
	}
	fmt.Println(res.Text())
	// Output: <!DOCTYPE html>...
}
```
### Configuration options
When you require further customization of the request, you can do so by specifying configuration `Options`:
```go
package main

import (
	"fmt"
	"github.com/sleeyax/gotcha"
	"log"
)

func main() {
	res, _ := gotcha.Post("https://httpbin.org/anything", &gotcha.Options{
		Json: gotcha.JSON{
			"hello": "world",
		},
		FollowRedirect: false,
	})
	body, _ := res.Json()
	defer res.Close()
    fmt.Println(body["data"])
	// Output: {"hello": "world"}
}
```
### Client
For advanced requests, create a client instance.
Clients are configurable, extendable & reusable. This gives you fine-grained control over the request:
```go
package main

import (
	"fmt"
	"github.com/sleeyax/gotcha"
	"io"
	"log"
	"net/http"
	"strings"
)

func main() {
	client, _ := gotcha.NewClient(&gotcha.Options{
		PrefixURL: "https://httpbin.org/",
		Headers: http.Header{
			"user-agent": {"gotcha by Sleeyax (https://github.com/sleeyax/gotcha)"},
		},
	})

	logClient, _ := client.Extend(&gotcha.Options{
		Hooks: gotcha.Hooks{
			Init: []gotcha.InitHook{
				func(o *gotcha.Options) {
					fmt.Println(fmt.Sprintf("about to send a request to %s with method %s", o.FullUrl.String(), o.Method))
				},
			},
		},
	})

	res, _ := logClient.DoRequest("PUT", "anything", &gotcha.Options{
		Body: io.NopCloser(strings.NewReader("hello world!")),
	})
	defer res.Close()
	// Output: "about to send a request to https://httpbin.org/anything with method PUT"
}
```

See [examples](examples) for more adapter-specific examples.
