# cclient adapter
This example contains a `cclient` package which implements an adapter for [cclient](https://github.com/x04/cclient).

## Usage
```go
package main

import (
	tls "github.com/refraction-networking/utls"
	"github.com/sleeyax/gotcha"
	"github.com/sleeyax/gotcha/examples/cclient/cclient"
)

func main() {
	adapter := cclient.NewAdapter(tls.HelloChrome_Auto)
	
	client, _ := gotcha.NewClient(&gotcha.Options{
		Adapter: adapter,
	})
	resp, err := client.Get("https://example.com")
	// ...

	// change TLS client hello ID at runtime
	adapter.ClientHello = tls.HelloFirefox_Auto
	resp, err = client.Get("https://example.com")
	// ...
}
```

## Test
`go test ./cclient`

`go run ./main.go`
