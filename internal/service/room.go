package service

import (
	"demo/internal/server/websocket"
	"fmt"
	"github.com/go-eagle/eagle/pkg/log"
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
	Cards       []uint8            // 当前轮次洗牌结果，记录索引
	CardsStatus []uint8            // 记录对应索引卡牌是否可用 0 -- 可以出 1 -- 已经出过 2 -- 扣下
	Scores      []int              // 每个座位的分数
	FirstPlayer *Player            // 第一个出牌的(头牌)
	CurrPlayer  *Player            // 当前出牌玩家
	userLock    sync.RWMutex       // 玩家相关操作的锁
	LastCard    Card               // 当前能出的牌 以此为标准， 头是第一张未用的，尾是最后一张未用的点数
	TableCards  []Card             // 桌上已经出的牌
}

func NewRoom(masterId string, roomId string) *Room {
	var users = make(map[string]*Player)
	player, _ := Manager.GetPlayerById(masterId)
	player.Seat = Seat1
	users[masterId] = player
	seat2Player := make([]*Player, TotalPlayers)
	seat2Player[0] = player
	room := &Room{
		RoomId:      roomId, // 用随机数生成，需要确保不重复
		RoomStatus:  GameReadying,
		Users:       users,
		Master:      masterId,
		Cards:       make([]uint8, TotalCardsCnt),
		CardsStatus: make([]uint8, TotalCardsCnt),
		Seat2Player: seat2Player,
		Scores:      make([]int, TotalPlayers),
		TableCards:  make([]Card, 0),
		LastCard:    InvalidCard,
	}
	player.Room = room
	return room
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
	if len(r.Users) != 4 {
		return false
	}
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
	defer r.userLock.RUnlock()
	player, ok := r.Users[userId]
	if !ok {
		err = fmt.Errorf("room no such user")
		return
	}
	return
}

func (r *Room) GameStart() (err error) {
	r.RoomStatus = GamePlaying
	for i, _ := range r.CardsStatus {
		r.CardsStatus[i] = 0
	}
	r.Cards = Shuffle()
	// 随机一个人出牌
	rand.Seed(time.Now().Unix())
	randomNumber := rand.Intn(100000)
	player := r.getUserBySeat(uint8(randomNumber % 4))
	if player == nil {
		return fmt.Errorf("internal error")
	}
	r.CurrPlayer = player
	r.FirstPlayer = player
	return err
}

func (r *Room) ResetGameAfterFinish() {
	for i := range r.Cards {
		r.Cards[i] = TotalCardsCnt
	}
	for i := range r.CardsStatus {
		r.Cards[i] = 0
	}
	r.LastCard = InvalidCard
	r.TableCards = make([]Card, 0)
	r.userLock.Lock()
	defer r.userLock.Unlock()
	for _, user := range r.Users {
		user.IsReady = false
		user.HandCardsGetted = false
	}
	log.Info("clear finished")
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
	player.Room = nil
	r.Seat2Player[player.Seat] = nil
	delete(r.Users, userId)
}

func (r *Room) getUserBySeat(seat uint8) *Player {
	r.userLock.RLock()
	defer r.userLock.RUnlock()
	player := r.Seat2Player[seat]
	return player
}

func (r *Room) DisableCard(card Card, seat uint8) (err error) {
	if r.LastCard == InvalidCard {
		err = fmt.Errorf("cannot disable card")
		return
	}

	if r.currSeatHavePlayableCard(r.LastCard, r.getCardsBySeat(seat), seat) {
		err = fmt.Errorf("have cards to play")
		return
	}

	idx, _ := r.getCardIdx(card, seat)
	// 扣牌
	r.CardsStatus[seat*HandCardCount+idx] = 2
	r.CurrPlayer = r.getUserBySeat((seat + 1) % 4)
	return
}

func (r *Room) PlayWithoutChooseHead(card Card, seat uint8) (isFinish bool, needChoose bool, err error) {
	playable := r.isCardPlayable(card, seat)
	log.Info("play card entered, playable:", playable, " card:", card)
	isFinish = false
	needChoose = false
	if playable {
		// 出牌逻辑，主要是把那张牌置空
		idx, cardIdx := r.getCardIdx(card, seat)
		log.Info("cardIdx: ", cardIdx)
		// 错误，出了不存在的牌
		if cardIdx == TotalCardsCnt {
			err = fmt.Errorf("no such card")
			return
		}
		// 对应卡牌状态置为出掉
		// 能出牌的场景
		r.CardsStatus[seat*HandCardCount+idx] = 1
		// 判断是否需要选择 出到头部/尾部 (可以出在头部 并且 可以出在尾部)
		if r.LastCard != InvalidCard {
			if checkNeedChoose(r.LastCard, card) {
				// 需要进行选择
				log.Info("need choose")
				needChoose = true
				return
			}
		}

		// 更新当前卡牌状态
		if r.LastCard == InvalidCard {
			r.LastCard = card
			r.TableCards = append(r.TableCards, card)
		} else {
			if r.LastCard.Head == card.Head {
				r.LastCard.Head = card.Tail
				r.TableCards = insertAtBeginning(r.TableCards, card)
			} else if r.LastCard.Head == card.Tail {
				r.LastCard.Head = card.Head
				r.TableCards = insertAtBeginning(r.TableCards, card)
			} else if r.LastCard.Tail == card.Head {
				r.LastCard.Tail = card.Tail
				r.TableCards = append(r.TableCards, card)
			} else if r.LastCard.Tail == card.Tail {
				r.LastCard.Tail = card.Head
				r.TableCards = append(r.TableCards, card)
			}
		}

		r.CurrPlayer = r.getUserBySeat((seat + 1) % 4)
		// 检查是否要算账/ 牌是否出完
		if r.CheckIfNeedFinish(seat) || r.CheckIfNeedSettle(card, seat) {
			isFinish = true
			return
		}
	} else {
		err = fmt.Errorf("cannot play this card")
	}
	return
}

func (r *Room) PlayWithChooseHead(card Card, onHead bool, seat uint8) (isFinish bool, err error) {
	log.Info("in playwitchooseHead : ", card, " ", onHead)
	if onHead {
		r.TableCards = insertAtBeginning(r.TableCards, card)
		if r.LastCard.Head == card.Head {
			r.LastCard.Head = card.Tail
		} else {
			r.LastCard.Head = card.Head
		}
	} else {
		r.TableCards = append(r.TableCards, card)
		if r.LastCard.Tail == card.Head {
			r.LastCard.Tail = card.Tail
		} else if r.LastCard.Tail == card.Tail {
			r.LastCard.Tail = card.Head
		}
	}
	r.CurrPlayer = r.getUserBySeat((seat + 1) % 4)
	// 检查是否要算账/ 牌是否出完
	if r.CheckIfNeedFinish(seat) || r.CheckIfNeedSettle(card, seat) {
		isFinish = true
		return
	}
	return false, nil
}

func (r *Room) getCardIdx(card Card, seat uint8) (resIdx uint8, resVal uint8) {
	cards := r.getCardsBySeat(seat)
	resVal = TotalCardsCnt
	for idx, i := range cards {
		if r.CardsStatus[idx+int(seat*HandCardCount)] == 0 && card == AllCards[i] {
			resIdx = uint8(idx)
			resVal = i
		}
	}
	return
}

// 获取对应座位的所有手牌，如果需要获得卡牌的原始索引需要 + seat*HandCardCount
func (r *Room) getCardsBySeat(seat uint8) []uint8 {
	cards := make([]uint8, HandCardCount)
	cards = r.Cards[seat*HandCardCount : seat*HandCardCount+HandCardCount]
	return cards
}

func (r *Room) isCardPlayable(card Card, seat uint8) bool {
	var playablePlayers = 0
	if r.LastCard == InvalidCard {
		// 第一次出牌， 不能拉三家，即出的牌不能是其他三家都没有竖牌的牌；
		_, cardIdx := r.getCardIdx(card, seat)
		// 没找到能出的牌， 出错，不应该有这种case
		if cardIdx == TotalCardsCnt {
			return false
		}
		otherSeat := (seat + 1) % TotalPlayers
		for otherSeat != seat {
			if r.currSeatHavePlayableCard(card, r.getCardsBySeat(otherSeat), otherSeat) {
				playablePlayers += 1
			}
			otherSeat = (otherSeat + 1) % TotalPlayers
		}
		if playablePlayers == 0 {
			// 不能其他三家没有牌出
			return false
		}
		return true
	} else {
		return checkCardCanPlay(r.LastCard, card, false)
	}
}

func (r *Room) CheckIfNeedSettle(card Card, seat uint8) bool {
	// 1 --- 检查是否需要算账
	var playablePlayers = 0
	log.Info("check if need settle, seat:", seat, "card:", card)
	otherSeat := (seat + 1) % TotalPlayers
	for otherSeat != seat {
		if r.currSeatHavePlayableCard(card, r.getCardsBySeat(otherSeat), otherSeat) {
			playablePlayers += 1
		}
		otherSeat = (otherSeat + 1) % TotalPlayers
	}
	if playablePlayers != 0 {
		log.Info("do not need settle")
		return false
	}
	// 优先级从高到低：
	// TODO: 1. 打牌的时候只能两头上, 需要check是否要两头上 && 增加选择流程 done

	// TODO: 2. 需要增加算账时候把能出的横牌扣掉的逻辑 done

	// TODO: 3. 结算时显示手牌+分数

	// TODO: 4. 显示手牌数

	// TODO: - 5 优先级较低 发牌规则做到前端
	// 分数排名从小到大
	log.Info("need settle")
	if seat == r.FirstPlayer.Seat {
		// 头牌算账
		log.Info("--first player settle--")
		cardCnt := 0
		idx := seat * HandCardCount
		for idx < seat*HandCardCount+HandCardCount {
			if r.CardsStatus[idx] == 0 {
				cardCnt += 1
			}
			idx++
		}
		if cardCnt == 1 {
			r.calcScoreNorMal()
			log.Info("--first player normal score --")
		} else {
			// 判断是否算账成功
			r.calcSettled(seat)
			log.Info("--first player settle score --")
		}
	} else if !r.currSeatHavePlayableCard(r.LastCard, r.getCardsBySeat(seat), seat) {
		// 是否是死砸账
		r.calcScoreNorMal()
	} else {
		// 普通算分
		r.calcSettled(seat)
	}
	log.Info("settled")
	return true
}

// 当前用户是否出完牌，只要第一个人出完牌就是赢家
func (r *Room) CheckIfNeedFinish(seat uint8) bool {
	// 是否所有人牌都出完
	log.Info("check if need finish, seat: ", seat)
	idx := seat * HandCardCount
	for idx < seat*HandCardCount+HandCardCount {
		if r.CardsStatus[idx] == 0 {
			log.Info("do not need finish")
			return false
		}
		idx++
	}
	r.calcScoreNorMal()
	return true
}

// 返回数组，第一个是手牌点数最小的
func (r *Room) calcCounts() []int {
	var seat uint8 = 0
	counts := make([]uint8, 4)
	log.Info("enter calcCounts")
	for seat < TotalSeats {
		idx := seat * HandCardCount
		for idx < seat*HandCardCount+HandCardCount {
			i := r.Cards[idx]
			if !AllCards[i].isStanding() && (AllCards[i].Head == r.LastCard.Head || AllCards[i].Head == r.LastCard.Tail) {
				// 算分时候去除能出的横牌
				r.CardsStatus[idx] = 1
			}
			if r.CardsStatus[idx] != 1 {
				counts[seat] += AllCards[i].GetCount()
			}
			idx++
		}
		seat++
	}
	log.Info("finish calcCounts: ", counts)
	return sort(counts)
}

func (r *Room) calcScoreNorMal() {
	log.Info("enter normal calc score")
	seats := r.calcCounts()
	r.Scores[seats[0]] += 6
	r.Scores[seats[1]] -= 1
	r.Scores[seats[2]] -= 2
	r.Scores[seats[3]] -= 3
}

func (r *Room) calcSettled(seat uint8) {
	counts := r.calcCounts()
	log.Info("calc settled: ", counts)
	if counts[0] == int(seat) {
		r.Scores[seat] += 12
		r.Scores[counts[1]] -= 2
		r.Scores[counts[2]] -= 4
		r.Scores[counts[3]] -= 6
	} else {
		r.Scores[seat] -= 12
		r.Scores[counts[0]] += 12
	}
}

// 检查当前座位是否有竖牌可以出
func (r *Room) currSeatHavePlayableCard(lastCard Card, cards []uint8, seat uint8) bool {
	for idx, card := range cards {
		if r.CardsStatus[idx+int(seat*HandCardCount)] != 0 {
			continue
		}
		if checkCardCanPlay(lastCard, AllCards[card], true) {
			return true
		}
	}
	return false
}

// 检查是否有竖牌可以出
func checkCardCanPlay(lastCard Card, currCard Card, withStanding bool) bool {
	if lastCard.Head == currCard.Head || lastCard.Tail == currCard.Tail || lastCard.Tail == currCard.Head || lastCard.Head == currCard.Tail {
		if (withStanding && currCard.isStanding()) || !withStanding {
			log.Info("can play lastCard: ", lastCard, "currCard: ", currCard, "withStanding: ", withStanding)
			return true
		} else {
			log.Info("can not paly lastCard: ", lastCard, "currCard: ", currCard, "withStanding: ", withStanding)
			return false
		}
	}
	return false
}

func sort(array []uint8) []int {
	var n int = len(array)
	var index []int
	for i := 0; i < n; i++ {
		index = append(index, i)
	}
	// fmt.Println(index)
	// fmt.Println("数组array的长度为：", n)
	if n < 2 {
		return nil
	}
	for i := 1; i < n; i++ {
		// fmt.Printf("检查第%d个元素%f\t", i, array[i])
		var temp = array[i]
		var tempIndex = index[i]
		var k int = i - 1
		for k >= 0 && array[k] > temp {
			k--
		}
		for j := i; j > k+1; j-- {
			array[j] = array[j-1]
			index[j] = index[j-1]
		}
		// fmt.Printf("其位置为%d\n", k+1)
		array[k+1] = temp
		index[k+1] = tempIndex
	}
	return index
}

func insertAtBeginning(arr []Card, element Card) []Card {
	newArr := make([]Card, len(arr)+1) // 定义新数组
	newArr[0] = element                // 将元素插入第一个位置
	copy(newArr[1:], arr)              // 将原数组中的元素往后移动一位
	return newArr                      // 返回新数组
}

func checkNeedChoose(lastCard Card, card Card) bool {
	if (lastCard.Head == card.Head || lastCard.Head == card.Tail) && (lastCard.Tail == card.Head || lastCard.Tail == card.Tail) {
		return true
	}
	return false
}
