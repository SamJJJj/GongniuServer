package server

import (
	"demo/internal/routers"
	"demo/internal/server/websocket"
	"github.com/go-eagle/eagle/pkg/app"
)

func NewWSServer(c *app.ServerConfig) *websocket.WsServer {
	server := websocket.NewWsServer(
		websocket.WithAddress(c.Addr),
	)
	routers.RegisterWSRouter()
	return server
}
