package websocket

import (
	"context"
	"log"
)

type WsServer struct {
	*WsServer
	path    string
	address string
}

func defaultWsServer() *WsServer {
	return &WsServer{
		address: ":8090",
		path:    "/ws",
	}
}

func NewWsServer(opts ...ServerOption) *WsServer {
	srv := defaultWsServer()
	// option 进行设置，需要完善
	for _, o := range opts {
		o(srv)
	}
	return srv
}

func (s *WsServer) Start(ctx context.Context) error {
	go StartWebSocket(s.path, s.address)
	log.Printf("[websocket] server is listening on: ", s.address, s.path)
	return nil
}

func (s *WsServer) Stop(ctx context.Context) error {
	return nil
}
