package server

import (
	"encoding/json"
	"entry-rpc/internal/codec"
	"entry-rpc/internal/common"
	"io"
	"log"
	"net"
	"sync"
)

type Server struct {
	serviceMap sync.Map
}

func NewServer() *Server {
	return &Server{}
}

var DefaultServer = NewServer()

func (server *Server) Accept(lis net.Listener) {
	// accept conn forever
	for {
		conn, err := lis.Accept()
		if err != nil {
			log.Println("rpc server: accept error:", err)
			return
		}
		// go for ServeConn
		go server.ServeConn(conn)
	}
}

func Accept(lis net.Listener) {
	DefaultServer.Accept(lis)
}

func (server *Server) ServeConn(conn io.ReadWriteCloser) {
	defer func() { _ = conn.Close() }()
	var opts common.Options
	// decode json options
	err := json.NewDecoder(conn).Decode(&opts)
	if err != nil {
		log.Printf("[WARN] options json parse error:%v\n", err)
		return
	}
	// opts valid
	if opts.Protocol != common.ProtocolName {
		log.Printf("[WARN] ProtocolName Unknown:%v\n", err)
		return
	}
	if opts.CodecType != common.CodecTypeName {
		log.Printf("[WARN] CodecType Unknown:%v\n", err)
		return
	}
	// serve codec
	mycodec := codec.NewGobCodec(conn)
	server.ServeCodec(mycodec)
}

var invalidRequest = struct{}{}

func (server *Server) ServeCodec(c *codec.GobCodec) {
	sending := new(sync.Mutex)
	wg := new(sync.WaitGroup)
	for {
		req, err := server.ReadRequest(c)
		//TODO: Handle ERR here
		if err != nil {
			if req == nil {
				break // it's not possible to recover, so close the connection
			}
			req.h.Err = err.Error()
			server.SendResponse(c, req.h, invalidRequest, sending)
			continue
		}
		wg.Add(1)
		go server.HandleRequest(c, req, wg, sending)
	}
	wg.Wait()
	_ = c.Close()
}
