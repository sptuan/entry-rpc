package codec

import "io"

type Header struct {
	ServiceMethod string // example: "ServiceNamespace.Method"
	Seq           uint64 // request seq
	Err           string // error string
}

type Codec interface {
	io.Closer
	ReadHeader(*Header) error
	ReadBody(interface{}) error
	Write(*Header, interface{}) error
}
