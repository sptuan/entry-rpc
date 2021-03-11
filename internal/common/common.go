package common

const (
	CodecTypeName = "gob"
	ProtocolName  = "entry-rpc-v1"
)

type Options struct {
	Protocol  string
	CodecType string
}
