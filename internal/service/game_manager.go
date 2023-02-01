package service

import "demo/internal/server/websocket"

var gameManager = &GameManager{}

type GameManager struct {
	Users map[string]*websocket.Client
	Rooms map[string]*Room
}

func NewGameManager() *GameManager {
	return &GameManager{}
}

// UserLogin 用户登录
func (g *GameManager) UserLogin(userId string, client *websocket.Client) {
	g.Users[userId] = client
}

// UserLogout 用户登出
func (g *GameManager) UserLogout(userId string) {
	delete(g.Users, userId)
}

func (g *GameManager) CreateRoom(userId string) {
	// 生成RoomId
	roomId := "testId"
	room := NewRoom(userId, roomId)
	g.Rooms[roomId] = room
}

func (g *GameManager) DestroyRoom(roomId string) {
	delete(g.Rooms, roomId)
}
