package erpc

import (
	"entry-rpc/pkg/client"
	"entry-rpc/pkg/server"
	"net"
)

type Clt = client.Client
type Srv = server.Server

func Register(rcvr interface{}) error {
	return server.Register(rcvr)
}

func Accept(lis net.Listener) {
	server.Accept(lis)
}

func Dial(network, address string) (clt *client.Client, err error) {
	return client.Dial(network, address)
}
