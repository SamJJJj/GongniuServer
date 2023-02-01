package service

import "fmt"

const (
	GameReadying = iota
	GamePlaying
	GameFinished
)

const (
	Seat1 = iota
	Seat2
	Seat3
	Seat4
	TotalSeats
)

type Room struct {
	RoomId     string            // 房间id
	RoomStatus uint32            // 房间状态
	Users      map[string]uint32 // 所有用户的id
	Master     string            // 房主的用户id, 庄家的id
}

func NewRoom(masterId string, roomId string) *Room {
	var users = *new(map[string]uint32)
	users[masterId] = Seat1
	return &Room{
		RoomId:     roomId, // 用随机数生成，需要确保不重复
		RoomStatus: GameReadying,
		Users:      users,
		Master:     masterId,
	}
}

func (r *Room) AddPlayer(userId string) error {
	userCnt := len(r.Users)
	if userCnt < TotalSeats {
		r.Users[userId] = uint32(userCnt)
		return nil
	}
	return fmt.Errorf("no enough seat")
}

func (r *Room) RemovePlayer(userId string) error {
	userCnt := len(r.Users)

	if userCnt == 0 {
		return fmt.Errorf("no users to leave")
	}

	if userCnt == 1 {
		gameManager.DestroyRoom(r.RoomId)
		return nil
	}

	if userId == r.Master {
		currMasterSeat := r.Users[userId]
		for k, v := range r.Users {
			if v == (currMasterSeat+1)%TotalSeats {
				r.Master = k
			}
		}
	}
	delete(r.Users, userId)
	return nil
}
