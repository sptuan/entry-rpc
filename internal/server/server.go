package server

import (
	"encoding/json"
	"entry-rpc/internal/codec"
	"fmt"
	"io"
	"log"
	"net"
	"reflect"
	"sync"
)

const (
	ProtocolName  = "entry-rpc-v1"
	CodecTypeName = "gob"
)

type Options struct {
	Protocol  string
	CodecType string
}

var DefaultOptions = &Options{
	Protocol:  ProtocolName,
	CodecType: CodecTypeName,
}

type Server struct{}

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
	var opts Options
	// decode json options
	err := json.NewDecoder(conn).Decode(&opts)
	if err != nil {
		log.Printf("[WARN] options json parse error:%v\n", err)
		return
	}
	// opts valid
	if opts.Protocol != ProtocolName {
		log.Printf("[WARN] ProtocolName Unknown:%v\n", err)
		return
	}
	if opts.CodecType != CodecTypeName {
		log.Printf("[WARN] CodecType Unknown:%v\n", err)
		return
	}
	// serve codec
	mycodec := codec.NewGobCodec(conn)
	server.ServeCodec(mycodec)
}

func (server *Server) ServeCodec(c *codec.GobCodec) {
	sending := new(sync.Mutex)
	wg := new(sync.WaitGroup)
	for {
		req, err := server.ReadRequest(c)
		//TODO: Handle ERR here
		if err != nil {
			break
		}
		wg.Add(1)
		go server.HandleRequest(c, req, wg, sending)
	}
	wg.Wait()
	_ = c.Close()
}

type request struct {
	h            *codec.Header // header of request
	argv, replyv reflect.Value // argv and replyv of request
}

func (server *Server) ReadRequestHeader(cc codec.Codec) (*codec.Header, error) {
	var h codec.Header
	if err := cc.ReadHeader(&h); err != nil {
		if err != io.EOF && err != io.ErrUnexpectedEOF {
			log.Println("rpc server: read header error:", err)
		}
		return nil, err
	}
	return &h, nil
}

func (server *Server) ReadRequest(cc codec.Codec) (*request, error) {
	h, err := server.ReadRequestHeader(cc)
	if err != nil {
		return nil, err
	}
	req := &request{h: h}
	// TODO: now we don't know the type of request argv
	// day 1, just suppose it's string
	req.argv = reflect.New(reflect.TypeOf(""))
	if err = cc.ReadBody(req.argv.Interface()); err != nil {
		log.Println("rpc server: read argv err:", err)
	}
	return req, nil
}

func (server *Server) SendResponse(cc codec.Codec, h *codec.Header, body interface{}, sending *sync.Mutex) {
	sending.Lock()
	defer sending.Unlock()
	if err := cc.Write(h, body); err != nil {
		log.Println("rpc server: write response error:", err)
	}
}

func (server *Server) HandleRequest(cc codec.Codec, req *request, wg *sync.WaitGroup, sending *sync.Mutex) {
	// TODO, should call registered rpc methods to get the right replyv
	// day 1, just print argv and send a hello message
	defer wg.Done()
	log.Println(req.h, req.argv.Elem())
	req.replyv = reflect.ValueOf(fmt.Sprintf("geerpc resp %d", req.h.Seq))
	server.SendResponse(cc, req.h, req.replyv.Interface(), sending)
}
