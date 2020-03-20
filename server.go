package main

import (
	"fmt"
	"log"
	"net/url"
	"os"
	"./protocol"
)

func main() {
	url, err := url.Parse(os.Args[1])
	if err != nil {
		log.Panicln("Invalid URL/Hostname format")
	}
	fmt.Println("Starting Server")
	protocol.Start(url)
}
