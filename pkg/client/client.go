package client

import (
	"encoding/json"
	"entry-rpc/pkg/codec"
	"entry-rpc/pkg/common"
	"errors"
	"fmt"
	"io"
	"log"
	"net"
	"sync"
)

type Client struct {
	cc       *codec.GobCodec
	opt      *common.Options
	sending  sync.Mutex
	header   codec.Header
	mu       sync.Mutex
	seq      uint64
	pending  map[uint64]*Call
	closing  bool
	shutdown bool
}

var _ io.Closer = (*Client)(nil)

var ErrShutdown = errors.New("Error shutdown.")

func (client *Client) Close() error {
	client.mu.Lock()
	defer client.mu.Unlock()
	if client.closing == true {
		return ErrShutdown
	}
	client.closing = true
	return client.cc.Close()
}

func (client *Client) IsAvailable() bool {
	client.mu.Lock()
	defer client.mu.Lock()
	return client.closing || client.shutdown
}

func (client *Client) Receive() {
	var err error
	for err == nil {
		var h codec.Header
		err = client.cc.ReadHeader(&h)
		if err != nil {
			break
		}
		call := client.RemoveCall(h.Seq)
		switch {
		case call == nil:
			// get response but not readable
			// maybe server has process the request
			err = client.cc.ReadBody(nil)
		case h.Err != "":
			// get a error pkg rpc
			call.Error = fmt.Errorf(h.Err)
			err = client.cc.ReadBody(nil)
			call.done()
		default:
			// get a reply
			err = client.cc.ReadBody(call.Returns)
			if err != nil {
				call.Error = errors.New("reading body " + err.Error())
			}
			call.done()
		}
	}
	client.TerminateCalls(err)
}

func NewClient(conn net.Conn, opt *common.Options) (*Client, error) {
	if err := json.NewEncoder(conn).Encode(opt); err != nil {
		log.Println("rpc client: options error: ", err)
		_ = conn.Close()
		return nil, err
	}
	return newClientCodec(codec.NewGobCodec(conn), opt), nil
}

func newClientCodec(cc *codec.GobCodec, opt *common.Options) *Client {
	client := &Client{
		seq:     1, // seq starts with 1, 0 means invalid call
		cc:      cc,
		opt:     opt,
		pending: make(map[uint64]*Call),
	}
	go client.Receive()
	return client
}

func newDefaultOptions() *common.Options {
	return &common.Options{
		Protocol:  common.ProtocolName,
		CodecType: common.CodecTypeName,
	}
}

// Dial connects to an RPC server at the specified network address
func Dial(network, address string) (client *Client, err error) {
	opt := newDefaultOptions()
	conn, err := net.Dial(network, address)
	if err != nil {
		return nil, err
	}
	// close the connection if client is nil
	defer func() {
		if client == nil {
			_ = conn.Close()
		}
	}()
	return NewClient(conn, opt)
}
