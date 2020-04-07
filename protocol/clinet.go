package protocol

import (
	"log"
)

var idCurrentServer string

type RpcClient int

type RpcSendReply struct {
	Success bool
	Leader string
}

func (t *RpcClient) Send(value int, reply *RpcSendReply) error {
	success, leader := handleClientRequest(value)
	reply.Success = success
	reply.Leader = leader
	return nil
}

func Send(value int) {
	var reply RpcSendReply
	err := servers[idCurrentServer].connection.Call("RpcClient.Send", value, &reply)
	if !reply.Success {
		log.Println("Leader changed to:", reply.Leader)
		idCurrentServer = reply.Leader
		Send(value)
	}
	if err != nil {
		log.Println("Unable to send:", err)
	}
}
