package main

import (
	"fmt"
	"github.com/sleeyax/gotcha"
	"github.com/sleeyax/gotcha/examples/default-http-client/fhttp"
	http "github.com/useflyent/fhttp"
	"log"
)

func main() {
	client, err := gotcha.NewClient(&gotcha.Options{
		Adapter: fhttp.NewAdapter(),
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
	fmt.Println(res.Text())
}
