<h1 align="center">
  <img width="150" src="docs/assets/logo.png" />
  <p>Gotcha</p>
</h1>

<h4 align="center">Customizable HTTP client</h4>

Gotcha is an alternative to Go's standard [http client](https://golang.org/src/net/http/client.go) implementation, 
with an API inspired by [got](https://github.com/sindresorhus/got).
It can interface with other HTTP libraries through an adapter.

Gotcha works fine with [fhttp](https://github.com/zMrKrabz/fhttp) & [cclient](https://github.com/x04/cclient). 
See the [examples](examples) for their respective adapter implementations. 

**Note: further documentation & examples are WIP (coming soon)**

## Usage
Basic example:
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
	fmt.Println(res.Body)
	// Output: <!DOCTYPE html>...
}
```
Basic example with options:
```go
package main

import (
	"fmt"
	"github.com/sleeyax/gotcha"
	"log"
)

func main() {
	// Note: from now on errors are omitted by _ to make the code snippets shorter
	res, _ := gotcha.Post("https://httpbin.org/anything", &gotcha.Options{
		Json: gotcha.JSON{
			"hello": "world",
		},
	})
	body, _ := res.Json()
	defer res.Close()
    fmt.Println(body["data"])
	// Output: {"hello": "world"}
}
```
Client example:

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

	logClient.DoRequest("PUT", "anything", &gotcha.Options{
		Body: io.NopCloser(strings.NewReader("hello world!")),
	})
	// Output: "about to send a request to https://httpbin.org/anything with method PUT"
}
```
