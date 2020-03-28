package protocol

import (
	"log"
)

var idCurrentServer string

type RpcClient int

func (t *RpcClient) Send(value int, reply *bool) error {
	*reply = handleClientRequest(value)
	return nil
}

func Send(value int) {
	var reply bool
	err := connectionsServers[idCurrentServer].connection.Call("RpcClient.Send", &value, &reply)
	if err != nil {
		log.Println("Unable to send: ", err)
	}
}
