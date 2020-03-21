package protocol

import (
	"log"
	"net/url"
)

func Send(value int) {
	log.Printf("Send: " + string(value))
}

func Connect(url *url.URL) bool {
	log.Println("Connecting to server: " + url.Host)
	return true
}