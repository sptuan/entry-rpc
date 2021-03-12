package erpc

import (
	"entry-rpc/internal/client"
	"entry-rpc/internal/server"
	"net"
)

func Register(rcvr interface{}) error {
	return server.Register(rcvr)
}

func Accept(lis net.Listener) {
	server.Accept(lis)
}

func Dial(network, address string) (clt *client.Client, err error) {
	return client.Dial(network, address)
}
