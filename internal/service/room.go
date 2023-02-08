package service

import (
	"demo/internal/server/websocket"
	"fmt"
	"math/rand"
	"sync"
	"time"
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
	RoomId      string             // 房间id
	RoomStatus  uint32             // 房间状态
	Users       map[string]*Player // 所有用户的id
	Master      string             // 房主的用户id, 庄家的id
	Seat2Player []*Player          // 座位到玩家的映射
	Cards       []uint8            // 当前轮次洗牌结果，记录索引 ---  出过的牌标记为24
	CurrPlayer  *Player            // 当前出牌玩家
	userLock    sync.RWMutex       // 玩家相关操作的锁
	LastCard    Card               // 上一张卡片
}

func NewRoom(masterId string, roomId string) *Room {
	var users = make(map[string]*Player)
	player, _ := Manager.GetPlayerById(masterId)
	player.Seat = Seat1
	users[masterId] = player
	return &Room{
		RoomId:      roomId, // 用随机数生成，需要确保不重复
		RoomStatus:  GameReadying,
		Users:       users,
		Master:      masterId,
		Cards:       make([]uint8, TotalCardsCnt),
		Seat2Player: make([]*Player, TotalPlayers),
		LastCard:    InvalidCard,
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
		r.addUser(userId, player)
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
	r.removeUser(userId)
	return nil
}

func (r *Room) GetNeedNotifyClients(player *Player) []*websocket.Client {
	var result []*websocket.Client
	r.userLock.RLock()
	defer r.userLock.RUnlock()
	for _, p := range r.Users {
		if p.UserInfo.UserId != player.UserInfo.UserId {
			client, _ := Manager.GetClientByUid(p.UserInfo.UserId)
			result = append(result, client)
		}
	}
	return result
}

func (r *Room) GetAllClients() []*websocket.Client {
	var result []*websocket.Client
	r.userLock.RLock()
	defer r.userLock.RUnlock()
	for _, p := range r.Users {
		client, _ := Manager.GetClientByUid(p.UserInfo.UserId)
		result = append(result, client)
	}
	return result
}

func (r *Room) GetAllPlayers() []*Player {
	var result []*Player
	r.userLock.RLock()
	defer r.userLock.RUnlock()
	for _, p := range r.Users {
		result = append(result, p)
	}
	return result
}

func (r *Room) CheckIfRoomNeedStart() bool {
	res := true
	r.userLock.RLock()
	defer r.userLock.RUnlock()
	for _, player := range r.Users {
		if !player.IsReady {
			res = false
			break
		}
	}
	return res
}
func (r *Room) GetPlayerById(userId string) (player *Player, err error) {
	r.userLock.RLock()
	player, ok := r.Users[userId]
	if !ok {
		err = fmt.Errorf("room no such user")
		return
	}
	return
}

func (r *Room) GameStart() (err error) {
	r.RoomStatus = GamePlaying
	r.Cards = Shuffle()
	// 随机一个人出牌
	rand.Seed(time.Now().Unix())
	randomNumber := rand.Intn(100000)
	player := r.getUserBySeat(uint8(randomNumber % 4))
	if player == nil {
		return fmt.Errorf("internal error")
	}
	r.CurrPlayer = player
	return err
}

func (r *Room) CheckNeedPlay() bool { // 返回是否需要开始出牌
	r.userLock.RLock()
	defer r.userLock.RUnlock()
	for _, player := range r.Seat2Player {
		if player == nil || player.HandCardsGetted == false {
			return false
		}
	}
	return true
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

func (r *Room) addUser(userId string, player *Player) {
	r.userLock.Lock()
	defer r.userLock.Unlock()
	r.Seat2Player[player.Seat] = player
	r.Users[userId] = player
}

func (r *Room) removeUser(userId string) {
	r.userLock.Lock()
	defer r.userLock.Unlock()
	player, _ := r.Users[userId]
	r.Seat2Player[player.Seat] = nil
	delete(r.Users, userId)
}

func (r *Room) getUserBySeat(seat uint8) *Player {
	r.userLock.RLock()
	defer r.userLock.RUnlock()
	player := r.Seat2Player[seat]
	return player
}

func (r *Room) PlayCard(card Card, seat uint8) (err error) {
	playable := r.isCardPlayable(card, seat)
	if playable {
		// 出牌逻辑，主要是把那张牌置空
		idx, cardIdx := r.getCardIdx(card, seat)
		if cardIdx == TotalCardsCnt {
			err = fmt.Errorf("no such card")
			return
		}
		r.Cards[seat*HandCardCount+idx] = TotalCardsCnt
		// 检查是否要算账/ 牌是否出完
	} else {
		err = fmt.Errorf("cannot play this card")
	}
	return
}

func (r *Room) getCardIdx(card Card, seat uint8) (resIdx uint8, resVal uint8) {
	cards := r.getCardsBySeat(seat)
	resVal = TotalCardsCnt
	for idx, i := range cards {
		if i != TotalCardsCnt && card == AllCards[i] {
			resIdx = uint8(idx)
			resVal = i
		}
	}
	return
}

func (r *Room) getCardsBySeat(seat uint8) []uint8 {
	cards := make([]uint8, HandCardCount)
	cards = r.Cards[seat*HandCardCount : seat*HandCardCount+HandCardCount]
	return cards
}

func (r *Room) isCardPlayable(card Card, seat uint8) bool {
	var playablePlayers = 0
	if r.LastCard == InvalidCard {
		// 第一次出牌
		_, cardIdx := r.getCardIdx(card, seat)
		// 没找到能出的牌
		if cardIdx == TotalCardsCnt {
			return false
		}
		otherSeat := (seat + 1) % TotalPlayers
		for otherSeat != seat {
			if currSeatHavePlayableCard(card, r.getCardsBySeat(otherSeat)) {
				playablePlayers += 1
			}
		}
		if playablePlayers == 0 {
			// 不能其他三家没有牌出
			return false
		}
		return true
	} else {
		return checkCardCanPlay(r.LastCard, card)
	}
}

func currSeatHavePlayableCard(lastCard Card, cards []uint8) bool {
	for _, card := range cards {
		if card == TotalCardsCnt {
			continue
		}
		if checkCardCanPlay(lastCard, AllCards[card]) {
			return true
		}
	}
	return false
}

func checkCardCanPlay(lastCard Card, currCard Card) bool {
	if lastCard.Head == currCard.Head || lastCard.Tail == currCard.Tail || lastCard.Tail == currCard.Head || lastCard.Head == currCard.Tail {
		return true
	}
	return false
}
