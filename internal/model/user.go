package model

import (
	"github.com/go-eagle/eagle/pkg/log"
)

type User struct {
	UserId    string `json:"user_id,omitempty"`    // 用户id
	AccountId string `json:"account_id"`           // 账号id
	NickName  string `json:"nick_name,omitempty"`  // 昵称
	AvatarUrl string `json:"avatar_url,omitempty"` // 头像
}

func QueryUserById(userId string) (user *User, ok bool) {
	user = new(User)
	DB.Where("user_id = ?", userId).First(&user)
	if len(user.NickName) != 0 {
		ok = true
		return
	}
	log.Info("query user:", ok)
	return
}

func QueryUserByName(userName string) (user *User, ok bool) {
	user = new(User)
	DB.Where("nick_name = ?", userName).First(&user)
	if len(user.NickName) != 0 {
		log.Info("user name Exist")
		ok = true
		return
	}
	return
}

func InsertUser(u *User) {
	log.Info("insert user:", u)
	DB.Create(u)
}
