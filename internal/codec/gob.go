package codec

import (
	"bufio"
	"encoding/gob"
	"io"
	"log"
)

type GobCodec struct {
	conn    io.ReadWriteCloser
	buf     *bufio.Writer
	encoder *gob.Encoder
	decoder *gob.Decoder
}

// TODO:design Codec interface here.
// unable to Design Codec Interface NOW, use GobCodec temp.
//var _ Codec = (*GobCodec)(nil)

func NewGobCodec(conn io.ReadWriteCloser) *GobCodec {
	b := bufio.NewWriter(conn)
	return &GobCodec{
		conn:    conn,
		buf:     b,
		encoder: gob.NewEncoder(conn),
		decoder: gob.NewDecoder(conn),
	}
}

func (c *GobCodec) ReadHeader(h *Header) error {
	return c.decoder.Decode(h)
}

func (c *GobCodec) ReadBody(i interface{}) error {
	return c.decoder.Decode(i)
}

func (c *GobCodec) Close() error {
	return c.conn.Close()
}

func (c *GobCodec) Write(h *Header, i interface{}) (err error) {
	defer func() {
		_ = c.buf.Flush()
		if err != nil {
			_ = c.Close()
		}
	}()
	err = c.encoder.Encode(h)
	if err != nil {
		log.Printf("[WARN] encode header failed: %v", err)
		return err
	}
	err = c.encoder.Encode(i)
	if err != nil {
		log.Printf("[WARN] encode body failed: %v", err)
		return err
	}
	return nil
}
