package service

import (
	"demo/internal/model"
	"demo/internal/server/websocket"
	"fmt"
	"github.com/go-eagle/eagle/pkg/log"
	"math/rand"
	"strconv"
	"sync"
	"time"
)

var Manager = NewGameManager()

type PlayerInfo struct {
	Client *websocket.Client
	Player *Player
}

type GameManager struct {
	Users    map[string]*PlayerInfo
	Rooms    map[string]*Room
	userLock sync.RWMutex
	roomLock sync.RWMutex
}

func NewGameManager() *GameManager {
	return &GameManager{
		Users: make(map[string]*PlayerInfo),
		Rooms: make(map[string]*Room),
	}
}

// UserLogin 用户登录
func (g *GameManager) UserLogin(userId string, client *websocket.Client, userInfo *model.User) {
	g.userLock.Lock()
	defer g.userLock.Unlock()
	g.Users[userId] = &PlayerInfo{
		Client: client,
		Player: NewPlayer(userInfo),
	}
}

// UserLogout 用户登出
func (g *GameManager) UserLogout(userId string) {
	g.userLock.Lock()
	defer g.userLock.Unlock()
	delete(g.Users, userId)
}

func (g *GameManager) CreateRoom(userId string) string {
	// 随机数生成
	roomId := genSixNum()
	_, ok := g.Rooms[roomId]
	for ok {
		roomId = genSixNum()
		_, ok = g.Rooms[roomId]
	}
	room := NewRoom(userId, roomId)
	g.roomLock.Lock()
	defer g.roomLock.Unlock()
	log.Info("room created:", roomId)
	g.Rooms[roomId] = room
	return roomId
}

func (g *GameManager) DestroyRoom(roomId string) {
	g.roomLock.Lock()
	defer g.roomLock.Unlock()
	log.Info("room destroyed:", roomId)
	delete(g.Rooms, roomId)
}

func (g *GameManager) GetClientByUid(userId string) (*websocket.Client, bool) {
	g.userLock.RLock()
	defer g.userLock.RUnlock()
	playerInfo, ok := g.Users[userId]
	return playerInfo.Client, ok
}

func (g *GameManager) GetPlayerById(userId string) (*Player, bool) {
	g.userLock.RLock()
	defer g.userLock.RUnlock()
	playerInfo, ok := g.Users[userId]
	return playerInfo.Player, ok
}

func (g *GameManager) GetRoomById(roomId string) (room *Room, err error) {
	g.roomLock.RLock()
	defer g.roomLock.RUnlock()
	room, ok := g.Rooms[roomId]
	if !ok {
		room = nil
		err = fmt.Errorf("no such room")
	}
	return
}

func genSixNum() string {
	rand.Seed(time.Now().Unix())
	randomNumber := rand.Intn(1000000)
	randomNumberString := strconv.Itoa(randomNumber)
	if len(randomNumberString) < 6 {
		randomNumberString = fmt.Sprintf("%06s", randomNumberString)
	}
	return randomNumberString
}
