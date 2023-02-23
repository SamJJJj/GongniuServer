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

	response := model.PlayerReadyResponse{}

	data, err = json.Marshal(&response)
	if err != nil {
		log.Error("player ready json marshal error", message)
		code = ecode.InternalError
		data = []byte(err.Error())
		return
	}
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
	startIdx := uint8(seat) * service.HandCardCount
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
	if room.CheckNeedPlay() {
		clients := room.GetAllClients()
		playResponse := model.GamePlayingNotify{
			CurrPlayingSeat: room.CurrPlayer.Seat,
			Cards:           nil,
		}
		data, err = json.Marshal(&playResponse)
		websocket.NotifyMessage(clients, NotifyGamePlaying, code, data)
	}
	data, err = json.Marshal(&response)
	if err != nil {
		log.Error("check get cards json marshal error", message)
		code = ecode.InternalError
		data = []byte(err.Error())
		return
	}
	return
}

func PlayCardHandler(client *websocket.Client, cmd string, message []byte) (code uint32, data interface{}) {
	request := &model.PlayCardRequest{}
	err := json.Unmarshal(message, request)
	if err != nil {
		log.Error("play card params error", message, err)
		code = ecode.ParamsError
		data = []byte("param unmarshal error")
		return
	}
	room, err := manager.GetRoomById(request.RoomId)
	if err != nil {
		log.Error("play card params error", message, err)
		code = ecode.ParamsError
		data = []byte(err.Error())
		return
	}
	card := service.Card{
		Head: request.Card.Head,
		Tail: request.Card.Tail,
	}
	needChoose := false
	isFinish := false
	// room + user check
	if request.OnHead == 0 {
		isFinish, needChoose, err = room.PlayWithoutChooseHead(card, request.Seat)
	} else {
		isFinish, err = room.PlayWithChooseHead(card, request.OnHead == 1, request.Seat)
	}
	if err != nil {
		log.Error("play card error", message, err)
		code = ecode.InternalError
		data = []byte(err.Error())
		return
	}

	code = ecode.Success

	if needChoose {
		response := model.PlayCardResponse{NeedChooseSide: needChoose}
		data, err = json.Marshal(&response)
		if err != nil {
			log.Error("check get cards json marshal error", message)
			code = ecode.InternalError
			data = []byte(err.Error())
			return
		}
		return
	}

	if isFinish {
		// 通知游戏结束
		var scores = make([]model.ScoreInfo, service.TotalPlayers)
		for idx, score := range room.Scores {
			scores[idx].Score = score
			scores[idx].Seat = idx
		}
		response := model.GameFinishNotify{Scores: scores}
		data, err = json.Marshal(&response)
		if err != nil {
			log.Error("check get cards json marshal error", message)
			code = ecode.InternalError
			data = []byte(err.Error())
			return
		}
		clients := room.GetAllClients()
		websocket.NotifyMessage(clients, NotifyGameFinished, code, data)
		room.ResetGameAfterFinish()
	} else {
		cards := make([]model.CardsInfo, 0)
		for _, card := range room.TableCards {
			cards = append(cards, model.CardsInfo{
				Head: card.Head,
				Tail: card.Tail,
			})
		}
		response := model.GamePlayingNotify{CurrPlayingSeat: room.CurrPlayer.Seat, Cards: cards}
		data, err = json.Marshal(&response)
		if err != nil {
			log.Error("check get cards json marshal error", message)
			code = ecode.InternalError
			data = []byte(err.Error())
			return
		}
		clients := room.GetAllClients()
		websocket.NotifyMessage(clients, NotifyGamePlaying, code, data)
	}

	response := model.PlayCardResponse{NeedChooseSide: false}
	data, err = json.Marshal(&response)
	if err != nil {
		log.Error("check get cards json marshal error", message)
		code = ecode.InternalError
		data = []byte(err.Error())
		return
	}
	return
}

func DisableCardHandler(client *websocket.Client, cmd string, message []byte) (code uint32, data interface{}) {
	request := &model.DisableCardRequest{}
	err := json.Unmarshal(message, request)
	if err != nil {
		log.Error("disable card params error", message, err)
		code = ecode.ParamsError
		data = []byte("param unmarshal error")
		return
	}
	room, err := manager.GetRoomById(request.RoomId)
	if err != nil {
		log.Error("disable card params error", message, err)
		code = ecode.ParamsError
		data = []byte(err.Error())
		return
	}
	card := service.Card{
		Head: request.Card.Head,
		Tail: request.Card.Tail,
	}
	// room + user check
	err = room.DisableCard(card, request.Seat)
	if err != nil {
		log.Error("disable card params error", message, err)
		code = ecode.ParamsError
		data = []byte(err.Error())
		return
	}

	cards := make([]model.CardsInfo, 0)
	for _, card := range room.TableCards {
		cards = append(cards, model.CardsInfo{
			Head: card.Head,
			Tail: card.Tail,
		})
	}
	response := model.GamePlayingNotify{CurrPlayingSeat: room.CurrPlayer.Seat, Cards: cards}
	data, err = json.Marshal(&response)
	if err != nil {
		log.Error("check get cards json marshal error", message)
		code = ecode.InternalError
		data = []byte(err.Error())
		return
	}
	clients := room.GetAllClients()
	websocket.NotifyMessage(clients, NotifyGamePlaying, code, data)

	code = ecode.Success
	response1 := model.DisableCardResponse{}
	data, err = json.Marshal(&response1)
	if err != nil {
		log.Error("disable card json marshal error", message)
		code = ecode.InternalError
		data = []byte(err.Error())
		return
	}
	return
}
