package handler

import (
	"demo/internal/ecode"
	"demo/internal/model"
	"demo/internal/server/websocket"
	"demo/internal/service"
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
	room, err := manager.GetRoomById(request.RoomId)
	if err != nil {
		code = ecode.InternalError
		data = []byte(err.Error())
		return
	}
	var players = *new([]model.PlayerInfo)
	for _, v := range room.GetAllPlayers() {
		players = append(players, model.PlayerInfo{
			User: model.UserInfo{
				NickName:  v.UserInfo.NickName,
				AvatarUrl: v.UserInfo.AvatarUrl,
			},
			Seat:    v.Seat,
			IsReady: v.IsReady,
		})
	}
	master, _ := manager.GetPlayerById(room.Master)

	response := &model.RoomMemberChangeNotify{
		CurrentSeat: seat,
		Players:     players,
		MasterSeat:  master.Seat,
	}

	data, err = json.Marshal(&response)
	if err != nil {
		log.Error("join room json marshal error", message)
		code = ecode.InternalError
		data = []byte(err.Error())
		return
	}
	clients := room.GetNeedNotifyClients(player)
	websocket.NotifyMessage(clients, NotifyRoomMemChange, code, data)
	return
}

func LeaveRoomHander(client *websocket.Client, cmd string, message []byte) (code uint32, data interface{}) {
	request := &model.LeaveRoomRequest{}
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
	err = player.LeaveRoom()
	if err != nil {
		code = ecode.InternalError
		data = []byte(err.Error())
		return
	}
	room, err := manager.GetRoomById(player.Room.RoomId)
	if err != nil {
		code = ecode.InternalError
		data = []byte(err.Error())
		return
	}

	master, _ := manager.GetPlayerById(room.Master)

	var players = *new([]model.PlayerInfo)
	for _, v := range room.Users {
		players = append(players, model.PlayerInfo{
			User: model.UserInfo{
				NickName:  v.UserInfo.NickName,
				AvatarUrl: v.UserInfo.AvatarUrl,
			},
			Seat:    v.Seat,
			IsReady: v.IsReady,
		})
	}

	response := &model.RoomMemberChangeNotify{
		CurrentSeat: service.TotalSeats,
		Players:     players,
		MasterSeat:  master.Seat,
	}

	data, err = json.Marshal(&response)
	if err != nil {
		log.Error("leave room json marshal error", message)
		code = ecode.InternalError
		data = []byte(err.Error())
		return
	}
	clients := room.GetNeedNotifyClients(player)
	websocket.NotifyMessage(clients, NotifyRoomMemChange, code, data)
	return
}
