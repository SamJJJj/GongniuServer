package handler

// 定义所有命令
const (
	Login               = "login"                  //登录
	CreateRoom          = "create_room"            //创建房间
	JoinRoom            = "join_room"              //加入房间
	LeaveRoom           = "leave_room"             //离开房间
	PlayerReady         = "player_ready"           //玩家准备
	GetHandCards        = "get_hand_cards"         //获取手牌
	NotifyRoomMemChange = "notify_room_mem_change" //房间成员变化的通知
	NotifyGameStart     = "notify_game_start"      //通知游戏开始
)
