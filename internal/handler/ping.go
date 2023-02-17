package handler

import (
	"demo/internal/server/websocket"
	"github.com/gin-gonic/gin"
	"time"

	"github.com/go-eagle/eagle/pkg/app"
	"github.com/go-eagle/eagle/pkg/log"
)

// Ping ping
// @Summary ping
// @Description ping
// @Tags system
// @Accept  json
// @Produce  json
// @Router /ping [get]
func Ping(c *gin.Context) {
	log.Info("Get function called.")

	app.Success(c, gin.H{})
}

func HeartbeatHandler(client *websocket.Client, cmd string, message []byte) (code uint32, data interface{}) {
	currTime := uint64(time.Now().Unix())
	client.Heartbeat(currTime)
	return
}
