package handler

import (
	"demo/internal/ecode"
	"demo/internal/model"
	"demo/internal/server/websocket"
	"demo/internal/service"
	"encoding/json"
	"github.com/go-eagle/eagle/pkg/log"
)

var manager = service.Manager

func LoginHandler(client *websocket.Client, cmd string, message []byte) (code uint32, data interface{}) {
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
	// service error 未处理
	manager.UserLogin(user.UserId, client, user)

	data, err = json.Marshal(&response)
	if err != nil {
		log.Error("login json marshal error", message)
		code = ecode.InternalError
		data = []byte("login error")
		return
	}
	return
}
