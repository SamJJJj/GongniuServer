package model

// 通用请求数据格式
type Request struct {
	Cmd  string      `json:"cmd"`            // 请求命令字
	Data interface{} `json:"data,omitempty"` // 数据 json
}

type LoginRequest struct {
	UserId    string `json:"user_id"`    // 用户id
	AccountId string `json:"account_id"` //账户id
}

type UserInfo struct {
	NickName  string `json:"nick_name"`
	AvatarUrl string `json:"avatar_url"`
}

type ScoreInfo struct {
	Score int `json:"score"`
	Seat  int `json:"seat"`
}

type CardsInfo struct {
	Head uint8 `json:"head"`
	Tail uint8 `json:"tail"`
}

type LoginResponse struct {
	User UserInfo `json:"user_info"`
}

type PlayerInfo struct {
	User    UserInfo `json:"user_info"`
	Seat    uint8    `json:"seat"`
	IsReady bool     `json:"is_ready"`
}

type CreateRoomRequest struct {
	UserId string `json:"user_id"`
}

type CreateRoomResponse struct {
	RoomId string `json:"room_id"`
}

type JoinRoomRequest struct {
	RoomId string `json:"room_id"`
	UserId string `json:"user_id"`
}

type LeaveRoomRequest struct {
	UserId string `json:"user_id"`
}

type PlayerReadyRequest struct {
	RoomId string `json:"room_id"`
	UserId string `json:"user_id"`
}

type PlayerReadyResponse struct {
}

type GameStartNotify struct {
}

type GetHandCardsRequest struct {
	UserId string `json:"user_id"`
	RoomId string `json:"room_id"`
	SeatNo string `json:"seat_no"`
}

type GetHandCardsResponse struct {
	Cards []CardsInfo `json:"cards"`
}

type CheckGetCardsRequest struct {
	UserId string `json:"user_id"`
	RoomId string `json:"room_id"`
}

type CheckGetCardsResponse struct {
}

type GamePlayingNotify struct {
	CurrPlayingSeat uint8       `json:"curr_playing_seat"`
	Cards           []CardsInfo `json:"curr_cards"`
}

type PlayCardRequest struct {
	UserId string    `json:"user_id"` // 是否能统一进行check?
	RoomId string    `json:"room_id"`
	Seat   uint8     `json:"seat"`
	Card   CardsInfo `json:"card"`
}

type PlayCardResponse struct {
}

type RoomMemberChangeNotify struct {
	CurrentSeat uint8        `json:"current_seat"`
	Players     []PlayerInfo `json:"players"`
	MasterSeat  uint8        `json:"master_seat"`
}

type GameFinishNotify struct {
	Scores []ScoreInfo `json:"scores"`
}

type DisableCardRequest struct {
	UserId string    `json:"user_id"` // 是否能统一进行check?
	RoomId string    `json:"room_id"`
	Seat   uint8     `json:"seat"`
	Card   CardsInfo `json:"card"`
}

type DisableCardResponse struct {
}
