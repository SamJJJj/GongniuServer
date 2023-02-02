package service

import (
	"demo/internal/server/websocket"
	"fmt"
	"sync"
)

const (
	GameReadying = iota
	GamePlaying
	GameFinished
)

const (
	Seat1 uint8 = iota
	Seat2
	Seat3
	Seat4
	TotalSeats
)

type Room struct {
	RoomId     string             // 房间id
	RoomStatus uint32             // 房间状态
	Users      map[string]*Player // 所有用户的id
	Master     string             // 房主的用户id, 庄家的id
	userLock   sync.RWMutex
}

func NewRoom(masterId string, roomId string) *Room {
	var users = make(map[string]*Player)
	player, _ := Manager.GetPlayerById(masterId)
	player.Seat = Seat1
	users[masterId] = player
	return &Room{
		RoomId:     roomId, // 用随机数生成，需要确保不重复
		RoomStatus: GameReadying,
		Users:      users,
		Master:     masterId,
	}
}

func (r *Room) AddPlayer(userId string) (err error, seat uint8) {
	userCnt := r.getUserLen()
	if userCnt == 0 {
		// 设置房主
		r.Master = userId
	}
	if userCnt < TotalSeats {
		player, _ := Manager.GetPlayerById(userId)
		seat = Seat1
		for !r.isSeatEmpty(seat) {
			seat = (seat + 1) % TotalSeats
		}
		player.Seat = seat
		r.userLock.Lock()
		defer r.userLock.Unlock()
		r.Users[userId] = player
		return
	}
	err = fmt.Errorf("no enough seat")
	seat = TotalSeats
	return
}

func (r *Room) RemovePlayer(userId string) error {
	userCnt := r.getUserLen()

	if userCnt == 0 {
		return fmt.Errorf("no users to leave")
	}

	if userCnt == 1 {
		Manager.DestroyRoom(r.RoomId)
		return nil
	}

	if userId == r.Master {
		currMasterSeat := r.getUserSeat(userId)
		currSeat := (currMasterSeat + 1) % TotalSeats
		for r.isSeatEmpty(currSeat) {
			currSeat = (currSeat + 1) % TotalSeats
		}
		for k, v := range r.Users {
			if v.Seat == currSeat {
				r.Master = k
			}
		}
	}
	delete(r.Users, userId)
	return nil
}

func (r *Room) GetNeedNotifyClients(player *Player) []*websocket.Client {
	var result []*websocket.Client
	for _, p := range r.Users {
		if p.UserInfo.UserId != player.UserInfo.UserId {
			client, _ := Manager.GetClientByUid(p.UserInfo.UserId)
			result = append(result, client)
		}
	}
	return result
}

func (r *Room) getUserLen() uint8 {
	r.userLock.RLock()
	defer r.userLock.RUnlock()
	return uint8(len(r.Users))
}

func (r *Room) getUserSeat(userId string) uint8 {
	r.userLock.RLock()
	defer r.userLock.RUnlock()
	return r.Users[userId].Seat
}

func (r *Room) isSeatEmpty(seat uint8) bool {
	for _, v := range r.Users {
		if v.Seat == seat {
			return false
		}
	}
	return true
}
