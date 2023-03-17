package websocket

import (
	"fmt"
	"github.com/go-eagle/eagle/pkg/log"
	"sync"
	"time"
)

type ClientManager struct {
	Clients       map[*Client]bool
	ClientsLock   sync.RWMutex
	Register      chan *Client
	Unregister    chan *Client
	Broadcast     chan []byte
	LogOutHandler func(client *Client)
}

func NewClientManager() (clientManager *ClientManager) {
	clientManager = &ClientManager{
		Clients:    make(map[*Client]bool),
		Register:   make(chan *Client, 1000),
		Unregister: make(chan *Client, 1000),
		Broadcast:  make(chan []byte, 1000),
	}
	return
}

func (manager *ClientManager) AddClient(client *Client) {
	manager.ClientsLock.Lock()
	defer manager.ClientsLock.Unlock()
	manager.Clients[client] = true
}

func (manager *ClientManager) DeleteClient(client *Client) (ok bool) {
	manager.ClientsLock.Lock()
	defer manager.ClientsLock.Unlock()

	if _, ok = manager.Clients[client]; ok {
		delete(manager.Clients, client)
		return
	}
	return
}

func (manager *ClientManager) RegisterEvent(client *Client) {
	manager.AddClient(client)
	log.Info("client connected", client.Addr)
}

func (manager *ClientManager) UnregisterEvent(client *Client) {
	manager.LogOutHandler(client)
	ok := manager.DeleteClient(client)
	if !ok {
		return
	}
	close(client.Send)
	log.Info("client disconnected", client.Addr)
}

func (manager *ClientManager) start() {
	for {
		select {
		case conn := <-manager.Register:
			manager.RegisterEvent(conn)
		case conn := <-manager.Unregister:
			manager.UnregisterEvent(conn)
		}
	}
}

func (manager *ClientManager) GetClients() (clients map[*Client]bool) {
	clients = make(map[*Client]bool)
	manager.ClientsLock.RLock()
	defer manager.ClientsLock.RUnlock()
	for key, value := range manager.Clients {
		clients[key] = value
	}
	return
}

func ClearTimeoutConnections() {
	currentTime := uint64(time.Now().Unix())

	clients := WebsocketManager.GetClients()
	for client := range clients {
		if client.IsHeartbeatTimeout(currentTime) {
			fmt.Println("心跳时间超时 关闭连接", client.Addr, client.HeartBeatTime)

			client.Socket.Close()
		}
	}
}
