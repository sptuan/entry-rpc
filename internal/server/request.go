package server

import (
	"entry-rpc/internal/codec"
	"entry-rpc/internal/service"
	"io"
	"log"
	"reflect"
	"sync"
)

type request struct {
	h            *codec.Header // header of request
	argv, replyv reflect.Value // argv and replyv of request
	mtype        *service.MethodType
	svc          *service.Service
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
	req.svc, req.mtype, err = server.findService(h.ServiceMethod)
	if err != nil {
		return req, err
	}
	req.argv = req.mtype.NewArgv()
	req.replyv = req.mtype.NewReplyv()

	// make sure that argvi is a pointer, ReadBody need a pointer as parameter
	argvi := req.argv.Interface()
	if req.argv.Type().Kind() != reflect.Ptr {
		argvi = req.argv.Addr().Interface()
	}
	if err = cc.ReadBody(argvi); err != nil {
		log.Println("rpc server: read body err:", err)
		return req, err
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
	defer wg.Done()
	err := req.svc.Call(req.mtype, req.argv, req.replyv)
	if err != nil {
		req.h.Err = err.Error()
		server.SendResponse(cc, req.h, invalidRequest, sending)
		return
	}
	server.SendResponse(cc, req.h, req.replyv.Interface(), sending)
}
