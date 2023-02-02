package websocket

import (
	"demo/internal/ecode"
	"demo/internal/model"
	"encoding/json"
	"fmt"
	"github.com/go-eagle/eagle/pkg/log"
	"sync"
)

type DisposeFunc func(client *Client, cmd string, message []byte) (code uint32, data interface{})

var (
	handlers        = make(map[string]DisposeFunc)
	handlersRWMutex sync.RWMutex
)

// 注册handler
func RegisterHandler(key string, handler DisposeFunc) {
	handlersRWMutex.Lock()
	defer handlersRWMutex.Unlock()
	handlers[key] = handler
}

func getHandler(key string) (handler DisposeFunc, ok bool) {
	handlersRWMutex.RLock()
	defer handlersRWMutex.RUnlock()
	handler, ok = handlers[key]
	return
}

func ProcessData(client *Client, message []byte) {
	log.Info("processing data", client.Addr, message)

	defer func() {
		if r := recover(); r != nil {
			fmt.Println("process stop", r)
		}
	}()

	request := &model.Request{}

	err := json.Unmarshal(message, request)
	if err != nil {
		log.Error("json marshal error", err)
		client.SendMsg([]byte("invalid request"))

		return
	}

	requestData, err := json.Marshal(request.Data)
	if err != nil {
		log.Error("process data json mashal", err)
		client.SendMsg([]byte("invalid request data"))

		return
	}
	cmd := request.Cmd

	var (
		code uint32
		msg  string
		data interface{}
	)

	if hander, ok := getHandler(cmd); ok {
		code, data = hander(client, cmd, requestData)
	} else {
		code = ecode.RoutNotExist
		log.Warn("no exist router", cmd)
	}
	msg = ecode.GetErrorMessage(code)
	responseHead := model.NewResponseHead(cmd, code, msg, data)

	headByte, err := json.Marshal(responseHead)
	if err != nil {
		log.Error("process head json", err)
		return
	}

	client.SendMsg(headByte)
	log.Info("send response to", client.Addr, "cmd", cmd, "code", code)
	return
}

func NotifyMessage(clients []*Client, cmd string, code uint32, data interface{}) {
	msg := ecode.GetErrorMessage(code)
	responseHead := model.NewResponseHead(cmd, code, msg, data)

	headByte, err := json.Marshal(responseHead)
	if err != nil {
		log.Error("notify head json", err)
		return
	}

	for _, c := range clients {
		c.SendMsg(headByte)
		log.Info("notifying msg to", c.Addr, "cmd", cmd, "code", code)
	}
	return
}
