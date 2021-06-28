package main

import (
	tls "github.com/refraction-networking/utls"
	"github.com/sleeyax/gotcha"
	"github.com/sleeyax/gotcha/examples/default-http-client/cclient"
	"log"
)

func main() {
	client, err := gotcha.NewClient(&gotcha.Options{
		Adapter:        cclient.NewAdapter(tls.HelloChrome_83),
		FollowRedirect: false,
	})
	if err != nil {
		log.Fatal(err)
	}

	resp, err := client.Get("https://example.com")
	if err != nil {
		log.Fatal(err)
	}
	defer resp.Close()

	log.Println(resp.Status)
}
