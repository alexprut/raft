package protocol

import (
	"log"
	"net/rpc"
)

var connectionsServers map[string]Server = make(map[string]Server)
var idCurrentServer string

type Server struct {
	url        string
	connection *rpc.Client
}

type RpcClient int

func (t *RpcClient) Send(value int, reply *bool) error {
	*reply = true
	log.Printf("Received: %d", value)
	return nil
}

func Send(value int) {
	var reply bool
	err := connectionsServers[idCurrentServer].connection.Call("RpcClient.Send", &value, &reply)
	if err != nil {
		log.Println("Unable to send: ", err)
	}
}

func Connect(url string) bool {
	log.Println("Connecting to server: " + url)
	connection, err := rpc.Dial("tcp", url)
	connectionsServers[url] = Server{url: url, connection: connection}
	if err != nil {
		log.Println("Unable Connecting to server: ", err)
		return false
	}
	idCurrentServer = url
	log.Println("Connected to server: " + url)
	return true
}
