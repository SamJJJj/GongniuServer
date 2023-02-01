package websocket

import "github.com/go-eagle/eagle/pkg/transport"

var _ transport.Server = (*WsServer)(nil)

// ServerOption is HTTP server option
type ServerOption func(*WsServer)

// WithAddress with server address.
func WithAddress(addr string) ServerOption {
	return func(s *WsServer) {
		//s.address = addr
	}
}
