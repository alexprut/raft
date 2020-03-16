package raft

import "fmt"

func send(value int) {
	fmt.Printf("Send: " + string(value))
}

func init() {
	fmt.Printf("Client Started")
}