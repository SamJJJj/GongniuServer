package handler

import (
	"demo/internal/ecode"
	"demo/internal/model"
	"demo/internal/server/websocket"
	"encoding/json"
	"github.com/go-eagle/eagle/pkg/log"
)

func PlayerReadyHandler(client *websocket.Client, cmd string, message []byte) (code uint32, data interface{}) {
	request := &model.PlayerReadyRequest{}
	err := json.Unmarshal(message, request)
	if err != nil {
		log.Error("player ready params error", message)
		code = ecode.ParamsError
		data = []byte("ready error")
		return
	}

	room, err := manager.GetRoomById(request.RoomId)
	if err != nil {
		log.Error("player ready params error", message)
		code = ecode.ParamsError
		data = []byte(err.Error())
		return
	}
	player, err := room.GetPlayerById(request.UserId)
	if err != nil {
		log.Error("player ready params error", message)
		code = ecode.ParamsError
		data = []byte(err.Error())
		return
	}
	player.Ready()

	code = ecode.Success
	response := model.PlayerReadyResponse{}

	data, err = json.Marshal(&response)
	if err != nil {
		log.Error("player ready json marshal error", message)
		code = ecode.InternalError
		data = []byte(err.Error())
		return
	}

	var (
		notifyCmd string
		clients   []*websocket.Client
	)

	if !room.CheckIfRoomNeedStart() {
		// 不需要开始游戏
		notifyCmd = NotifyRoomMemChange

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
			CurrentSeat: player.Seat,
			Players:     players,
			MasterSeat:  master.Seat,
		}
		data, err = json.Marshal(&response)
		if err != nil {
			log.Error("player ready json marshal error", message)
			code = ecode.InternalError
			data = []byte(err.Error())
			return
		}
	} else {
		// 开始游戏
		notifyCmd = NotifyGameStart
		response := model.GameStartNotify{}
		data, err = json.Marshal(&response)
		if err != nil {
			log.Error("player ready json marshal error", message)
			code = ecode.InternalError
			data = []byte(err.Error())
			return
		}
		room.GameStart()
	}
	websocket.NotifyMessage(clients, notifyCmd, code, data)

	return
}
