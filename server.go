package main

import (
	"./protocol"
	"log"
	"os"
)

// The first argument is the host:port of the server, e.g. "localhost:8000"
// the following arguments are the other peers
// run example: ./server localhost:8001 localhost:8002 localhost:8003 localhost:8004 localhost:8005
func main() {
	url := os.Args[1]
	log.Println("Starting Server")
	protocol.Start(url, os.Args[2:])
}
