package websocket

import (
	"fmt"
	"github.com/go-eagle/eagle/pkg/log"
	"github.com/gorilla/websocket"
	"net/http"
	"time"
)

var WebsocketManager = NewClientManager()

func StartWebSocket(path string, addr string) {
	http.HandleFunc(path, wsPage)
	http.ListenAndServe(addr, nil)
}

func wsPage(w http.ResponseWriter, req *http.Request) {
	conn, err := (&websocket.Upgrader{CheckOrigin: func(r *http.Request) bool {
		fmt.Println("升级协议", "ua:", r.Header["User-Agent"], "referer:", r.Header["Referer"])

		return true
	}}).Upgrade(w, req, nil)

	if err != nil {
		log.Error("init ws error : ", err)
		return
	}
	log.Info("webSocket connected:", conn.RemoteAddr().String())
	currentTime := uint64(time.Now().Unix())
	client := NewClient(conn.RemoteAddr().String(), conn, currentTime)
	client.connected()

	go client.write()
	go client.read()

	WebsocketManager.Register <- client
}
