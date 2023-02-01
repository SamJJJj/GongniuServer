package handler

import (
	"demo/internal/ecode"
	"demo/internal/model"
	"demo/internal/server/websocket"
	"encoding/json"
	"github.com/go-eagle/eagle/pkg/log"
)

func LoginHandler(client *websocket.Client, message []byte) (code uint32, data interface{}) {
	request := &model.LoginRequest{}
	err := json.Unmarshal(message, request)
	if err != nil {
		log.Error("login params error", message)
		code = ecode.ParamsError
		data = []byte("login error")
		return
	}
	// 查找是否有该用户/将用户写入数据库
	user, ok := model.QueryUserById(request.UserId)
	if !ok {
		// 查找到直接返回用户信息
		log.Info("register to mysql")
		user.UserId = request.UserId
		user.AccountId = request.AccountId
		user.NickName = "test"
		user.AvatarUrl = "test.png"
		model.InsertUser(user)
	}

	code = ecode.Success
	response := model.LoginResponse{
		User: *user,
	}
	data, err = json.Marshal(&response)
	if err != nil {
		log.Error("login params error", message)
		code = ecode.InternalError
		data = []byte("login error")
		return
	}
	return
}
