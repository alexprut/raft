package protocol

import "fmt"

func Send(value int) {
	fmt.Printf("Send: " + string(value))
}