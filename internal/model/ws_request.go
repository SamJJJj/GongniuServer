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

type LoginResponse struct {
	User User `json:"user_info"`
}

type PlayerInfo struct {
	User    User  `json:"user_info"`
	Seat    uint8 `json:"seat"`
	IsReady bool  `json:"is_ready"`
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

type RoomMemberChangeResponse struct {
	CurrentSeat uint8        `json:"current_seat"`
	Players     []PlayerInfo `json:"players"`
	MasterSeat  uint8        `json:"master_seat"`
}

type LeaveRoomRequest struct {
	UserId string `json:"user_id"`
}
