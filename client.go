package main

import (
	"./protocol"
	"log"
	"net/url"
	"os"
)

func main() {
	url, err := url.Parse(os.Args[1])
	if err != nil {
		log.Panicln("Invalid URL/Hostname format")
	}
	log.Printf("Started Client")
	protocol.Connect(url)
	protocol.Send(1) // FIXME remove example
}