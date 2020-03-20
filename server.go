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
	log.Println("Starting Server")
	protocol.Start(url)
}
