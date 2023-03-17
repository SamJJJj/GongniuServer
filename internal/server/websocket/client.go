package websocket

import (
	"github.com/go-eagle/eagle/pkg/log"
	"github.com/gorilla/websocket"
	"runtime/debug"
)

const (
	HearbeatExpirationTime = 6 * 60
)

type Client struct {
	Addr          string          // 客户端地址
	Socket        *websocket.Conn // 客户端连接对象
	Send          chan []byte     // 待发送数据
	FirstTime     uint64          // 首次连接时间
	HeartBeatTime uint64          // 上次心跳时间
}

func NewClient(addr string, socket *websocket.Conn, firstTime uint64) (client *Client) {
	client = &Client{
		Addr:          addr,
		Socket:        socket,
		Send:          make(chan []byte, 100), // 最多100条待发送
		FirstTime:     firstTime,
		HeartBeatTime: firstTime,
	}
	return
}

// 从客户端读数据
func (c *Client) read() {
	defer func() {
		if r := recover(); r != nil {
			log.Warn("write stop", string(debug.Stack()), r)
		}
	}()

	defer func() {
		log.Info("read client data & close send", c)
		close(c.Send)
	}()

	for {
		_, message, err := c.Socket.ReadMessage()
		if err != nil {
			log.Error("read client data error", c.Addr, err)
			return
		}

		// 处理读到的数据
		log.Info("processing client data", string(message))
		ProcessData(c, message)
	}
}

// 向客户端写数据
func (c *Client) write() {
	defer func() {
		if r := recover(); r != nil {
			log.Warn("write stop", string(debug.Stack()), r)
		}
	}()

	defer func() {
		WebsocketManager.Unregister <- c
		c.Socket.Close()
		log.Info("client send data finished", c)
	}()

	for {
		select {
		case message, ok := <-c.Send:
			if !ok {
				log.Error("get send data error", c.Addr)
				return
			}
			c.Socket.WriteMessage(websocket.TextMessage, message)
		}
	}
}
func (c *Client) SendMsg(msg []byte) {
	if c == nil {
		return
	}

	defer func() {
		if r := recover(); r != nil {
			log.Warn("write stop", string(debug.Stack()), r)
		}
	}()

	c.Send <- msg
}

func (c *Client) close() {
	close(c.Send)
}

func (c *Client) Heartbeat(currentTime uint64) {
	c.HeartBeatTime = currentTime
}

func (c *Client) IsHeartbeatTimeout(currentTime uint64) (timeout bool) {
	if c.HeartBeatTime+HearbeatExpirationTime <= currentTime {
		timeout = true
	}
	return
}

func (c *Client) connected() {
	WebsocketManager.AddClient(c)
}
