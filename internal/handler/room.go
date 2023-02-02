package handler

import (
	"demo/internal/ecode"
	"demo/internal/model"
	"demo/internal/server/websocket"
	"encoding/json"
	"github.com/go-eagle/eagle/pkg/log"
)

func CreatRoomHandler(client *websocket.Client, cmd string, message []byte) (code uint32, data interface{}) {
	request := &model.CreateRoomRequest{}
	err := json.Unmarshal(message, request)
	if err != nil {
		log.Error("create room params error", message)
		code = ecode.ParamsError
		data = []byte("create room error")
		return
	}
	// 未处理错误
	roomId := manager.CreateRoom(request.UserId)
	response := model.CreateRoomResponse{RoomId: roomId}
	code = ecode.Success
	data, err = json.Marshal(&response)
	if err != nil {
		log.Error("create room json marshal error", message)
		code = ecode.InternalError
		data = []byte("create room error")
		return
	}
	return
}

func JoinRoomHander(client *websocket.Client, cmd string, message []byte) (code uint32, data interface{}) {
	request := &model.JoinRoomRequest{}
	err := json.Unmarshal(message, request)
	if err != nil {
		log.Error("join room params error", message)
		code = ecode.ParamsError
		data = []byte("join room error")
		return
	}
	player, ok := manager.GetPlayerById(request.UserId)
	if !ok {
		log.Error("user id error", message)
		code = ecode.ParamsError
		data = []byte("user id error")
		return
	}
	err1, seat := player.JoinRoom(request.RoomId)
	if err1 != nil {
		code = ecode.InternalError
		data = []byte(err1.Error())
		return
	}
	var players = *new([]model.PlayerInfo)
	room, ok := manager.GetRoomById(request.RoomId)
	for _, v := range room.Users {
		players = append(players, model.PlayerInfo{
			User:    *v.UserInfo,
			Seat:    v.Seat,
			IsReady: v.IsReady,
		})
	}
	master, _ := manager.GetPlayerById(room.Master)

	response := &model.RoomMemberChangeResponse{
		CurrentSeat: seat,
		Players:     players,
		MasterSeat:  master.Seat,
	}

	data, err = json.Marshal(&response)
	if err != nil {
		log.Error("join room json marshal error", message)
		code = ecode.InternalError
		data = []byte("login error")
		return
	}
	clients := room.GetNeedNotifyClients(player)
	websocket.NotifyMessage(clients, cmd, code, data)
	return
}
