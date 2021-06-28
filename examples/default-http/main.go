package main

import (
	"fmt"
	"github.com/sleeyax/gotcha"
	"log"
)

func main() {
	res, err := gotcha.Get("https://sleeyax.dev", &gotcha.Options{
		Adapter: Adapter,
	})
	if err != nil {
		log.Fatal(err)
	}
	fmt.Println(res.Text())
	// Output: <!DOCTYPE html>...
}
