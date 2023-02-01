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
	User User `json: user`
}
