package main

import (
	"./protocol"
	"bufio"
	"log"
	"os"
	"strconv"
)

func main() {
	log.Printf("Started Client")
	for _, param := range os.Args[1:] {
		log.Printf("Registering Server: " + param)
		protocol.Connect(param)
	}
	input := bufio.NewScanner(os.Stdin)
	for input.Scan() {
		val, err := strconv.Atoi(input.Text())
		if err != nil {
			log.Printf("Invalid provided value")
		} else {
			protocol.Send(val)
		}
	}
}
