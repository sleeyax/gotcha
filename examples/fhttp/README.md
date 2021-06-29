# fhttp adapter
This example contains a `fhttp` package which implements an adapter for [fhttp](https://github.com/useflyent/fhttp).

## Usage

```go
package main

import (
	"github.com/sleeyax/gotcha"
	"github.com/sleeyax/gotcha/examples/default-http-client/fhttp"
	http "github.com/useflyent/fhttp"
	"log"
)

func main() {
	client, err := gotcha.NewClient(&gotcha.Options{
		// Adapter: fhttp.NewAdapter(),
		Adapter: &fhttp.Adapter{
			// Optionally define custom http.Transport options here 
			Transport: &http.Transport{},
		},
	})
	if err != nil {
		log.Fatal(err)
	}
	res, _ := client.Get("https://httpbin.org/anything", &gotcha.Options{
		Headers: map[string][]string{
			"user-agent":        {"gotcha By Sleeyax (https://github.com/sleeyax/gotcha)"},
			"foo":               {"bar"},
			"abc":               {"def", "123"},
			http.HeaderOrderKey: {"abc", "user-agent", "foo"},
		},
	})
	// ...
}
```

## Test
```shell
$ go run ./main.go
```
