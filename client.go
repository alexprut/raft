package main

import (
	"fmt"
	"./protocol"
)

func main() {
	fmt.Printf("Started Client")
	protocol.Send(1) // FIXME remove example
}