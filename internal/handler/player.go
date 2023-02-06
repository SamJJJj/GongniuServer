package handler

import (
	"demo/internal/ecode"
	"demo/internal/model"
	"demo/internal/server/websocket"
	"demo/internal/service"
	"encoding/json"
	"github.com/go-eagle/eagle/pkg/log"
	"strconv"
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

		clients = room.GetNeedNotifyClients(player)

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
		clients = room.GetAllClients()
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

func GetHandCardsHandler(client *websocket.Client, cmd string, message []byte) (code uint32, data interface{}) {
	request := &model.GetHandCardsRequest{}
	err := json.Unmarshal(message, request)
	if err != nil {
		log.Error("get cards params error", message)
		code = ecode.ParamsError
		data = []byte("param unmarshal error")
		return
	}

	room, err := manager.GetRoomById(request.RoomId)
	if err != nil {
		log.Error("get cards params error", message)
		code = ecode.ParamsError
		data = []byte(err.Error())
		return
	}
	// 检查用户是否在对应房间里
	_, err = room.GetPlayerById(request.UserId)
	if err != nil {
		log.Error("get cards params error", message)
		code = ecode.ParamsError
		data = []byte(err.Error())
		return
	}

	seat, err := strconv.Atoi(request.SeatNo)
	if err != nil {
		log.Error("get cards params error", message)
		code = ecode.ParamsError
		data = []byte(err.Error())
		return
	}
	startIdx := seat * service.HandCardCount
	cardsIdx := room.Cards[startIdx : startIdx+service.HandCardCount]
	cards := make([]model.CardsInfo, service.HandCardCount)
	for idx, i := range cardsIdx {
		cards[idx] = model.CardsInfo{
			Head: service.AllCards[i].Head,
			Tail: service.AllCards[i].Tail,
		}
	}

	code = ecode.Success
	response := model.GetHandCardsResponse{
		Cards: cards,
	}

	data, err = json.Marshal(&response)
	if err != nil {
		log.Error("get hand card json marshal error", message)
		code = ecode.InternalError
		data = []byte(err.Error())
		return
	}
	return
}

func CheckGetCardsHandler(client *websocket.Client, cmd string, message []byte) (code uint32, data interface{}) {
	request := &model.CheckGetCardsRequest{}
	err := json.Unmarshal(message, request)
	if err != nil {
		log.Error("check get cards  params error", message)
		code = ecode.ParamsError
		data = []byte("param unmarshal error")
		return
	}
	room, err := manager.GetRoomById(request.RoomId)
	if err != nil {
		log.Error("get cards params error", message, err)
		code = ecode.ParamsError
		data = []byte(err.Error())
		return
	}
	// 检查用户是否在对应房间里
	player, err := room.GetPlayerById(request.UserId)
	if err != nil {
		log.Error("get cards params error", message)
		code = ecode.ParamsError
		data = []byte(err.Error())
		return
	}
	player.CardsGetted()
	code = ecode.Success
	response := model.CheckGetCardsResponse{}

	data, err = json.Marshal(&response)
	if err != nil {
		log.Error("check get cards json marshal error", message)
		code = ecode.InternalError
		data = []byte(err.Error())
		return
	}
	return
}
