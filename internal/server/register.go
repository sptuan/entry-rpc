package server // Register publishes in the server the set of methods of the
import (
	"entry-rpc/internal/service"
	"errors"
)

func (server *Server) Register(rcvr interface{}) error {
	s := service.NewService(rcvr)
	if _, dup := server.serviceMap.LoadOrStore(s.GetName(), s); dup {
		return errors.New("rpc: service already defined: " + s.GetName())
	}
	return nil
}

// Register publishes the receiver's methods in the DefaultServer.
func Register(rcvr interface{}) error {
	return DefaultServer.Register(rcvr)
}
