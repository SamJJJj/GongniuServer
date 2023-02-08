package service

import (
	"demo/internal/model"
	"fmt"
)

type Player struct {
	UserInfo        *model.User
	Room            *Room
	IsReady         bool
	Seat            uint8
	HandCardsGetted bool
	HandCards       []Card
}

func NewPlayer(userInfo *model.User) *Player {
	return &Player{
		UserInfo:        userInfo,
		Seat:            TotalSeats,
		HandCardsGetted: false,
		IsReady:         false,
	}
}

func (p *Player) JoinRoom(roomId string) (err error, seat uint8) {
	room, ok := Manager.Rooms[roomId]
	if !ok {
		err = fmt.Errorf("no such room:", roomId)
		return
	}
	err, seat = room.AddPlayer(p.UserInfo.UserId)
	if err != nil {
		return
	}
	p.Room = room
	return
}

func (p *Player) LeaveRoom() error {
	if p.Room == nil {
		return fmt.Errorf("leave without a room", p.UserInfo.UserId)
	}
	err := p.Room.RemovePlayer(p.UserInfo.UserId)
	if err != nil {
		return err
	}
	p.Room = nil
	return nil
}

func (p *Player) Ready() {
	if p.IsReady {
		return
	}
	p.IsReady = true
}

func (p *Player) CardsGetted() {
	p.HandCardsGetted = true
}
