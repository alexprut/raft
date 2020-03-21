package main

import (
	"./protocol"
	"log"
	"os"
)

func main() {
	url := os.Args[1]
	log.Println("Starting Server")
	protocol.Start(url)
}
