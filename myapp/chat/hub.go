// Copyright 2013 The Gorilla WebSocket Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package main

import (
	"encoding/json"
	"fmt"
	"strconv"
	"strings"
	"sync"
)

var mu sync.RWMutex

// Hub maintains the set of active clients and broadcasts messages to the
// clients.
type Hub struct {
	// chat rooms
	rooms map[string]map[*Client]bool
	// Registered clients.
	clients map[*Client]bool

	// Inbound messages from the clients.
	broadcast chan []byte

	loadmsg chan *Client

	// Register requests from the clients.
	register chan *Client

	// Unregister requests from clients.
	unregister chan *Client
}

//預計再開一個chan用來處理redis存資料

type SysMsg struct {
	Text     string `json:"text"`
	RoomInfo string `json:"room_info"`
	UserInfo string `json:"user_info"`
}

func newHub() *Hub {
	return &Hub{
		rooms:      make(map[string]map[*Client]bool),
		broadcast:  make(chan []byte),
		loadmsg:    make(chan *Client),
		register:   make(chan *Client),
		unregister: make(chan *Client),
		clients:    make(map[*Client]bool),
	}
}

func (h *Hub) run() {

	for {
	FOR:
		select {
		/*註冊新的使用者
		* 根據客戶中的房號參數判斷
		* 是否要建立新聊天室房間
		* 產生系統資訊準備傳給前端
		 */
		case client := <-h.register:
			conns := h.rooms[client.roomId]

			/*如果聊天室不存在即建立新的*/
			mu.Lock() //鎖定
			if conns == nil {
				conns = make(map[*Client]bool)
				h.rooms[client.roomId] = conns //->寫
				fmt.Println("新的聊天室被創建了")

			}
			h.rooms[client.roomId][client] = true //將使用者存入聊天室map //->寫
			mu.Unlock()                           //解除鎖定
			fmt.Println("rooms：", h.rooms)

			/*判斷是否為私聊*/
			if client.roomType == "private" {
				break FOR
			}

			h.makeInfo()

			/*系統資訊：所在房間人員名單*/
			roomstate := make([]string, 0, len(conns))
			for con := range conns {
				roomstate = append(roomstate, con.id)
			}
			/*製作系統提示(訊息＋人員名單)*/
			sysmsg := client.id + " 進入 " + client.roomId + " 聊天室!"
			data, _ := json.Marshal(&SysMsg{Text: sysmsg, RoomInfo: client.roomId, UserInfo: strings.Join(roomstate, ",")})
			message, _ := json.Marshal(&Message{Sender: "SYS", RoomId: client.roomId, Type: "H", Content: string(data), Time: getUTCTime()})

			/*發送至該聊天室*/
			for con := range conns {
				con.send <- message
			}

		/*使用者離線或切換聊天室
		* 清除該客戶的連線資訊
		* 如果離開後聊天室為空，則關閉聊天室
		* 發送系統提示
		 */
		case client := <-h.unregister:
			conns := h.rooms[client.roomId]
			if conns != nil {
				if _, ok := conns[client]; ok {
					cleave := client.id //保留id以用做系統提示
					delete(conns, client)
					close(client.send)

					/*聊天室若為空，則刪除該聊天室*/
					if len(conns) == 0 {
						delete(h.rooms, client.roomId)
						h.makeInfo()
					}

					/*判斷是否為私聊*/
					if client.roomType == "private" {
						break FOR
					}

					/*系統資訊：所在房間人員名單*/
					roomstate := make([]string, 0, len(conns))
					for con := range conns {
						roomstate = append(roomstate, con.id)
					}
					/*製作系統提示(訊息＋人員名單)*/
					sysmsg := cleave + " 離開 " + client.roomId + " 聊天室!"
					data, _ := json.Marshal(&SysMsg{Text: sysmsg, RoomInfo: client.roomId, UserInfo: strings.Join(roomstate, ",")})
					message, _ := json.Marshal(&Message{Sender: "SYS", RoomId: client.roomId, Type: "H", Content: string(data), Time: getUTCTime()})

					/*發送至該聊天室*/
					for con := range conns {
						con.send <- message
					}
				}
			}
		/*廣播使用者輸入的訊息至聊天室
		* 跨聊天室廣播
		* 一般訊息發送
		 */
		case message := <-h.broadcast:
			var msg Message
			err := json.Unmarshal(message, &msg) //轉換出使用者發的訊息內容
			if err != nil {
				fmt.Println(message)
			}

			/*如果是廣播訊息，則發送至全頻道後跳轉回迴圈頂端*/
			if msg.Type == "A" {
				h.sys(message)
				break FOR
			}

			/*一般訊息只發送到該聊天室*/
			conns := h.rooms[msg.RoomId]
			for con := range conns {
				select {
				case con.send <- message:
				default:
					close(con.send)
					delete(h.rooms, msg.RoomId)
				}
			}

		//如果是私訊 只發給該使用者

		case client := <-h.loadmsg:
			user_room := client.roomId
			data := zrangeMessage(user_room, 100)
			/*印出歷史訊息*/
			for k := range data {
				msg := data[k].Member.(string)
				client.send <- []byte(msg)
			}

		}
	}
}

/*獲取聊天室列表和使用者名單*/
func (h *Hub) makeInfo() []byte {
	chatrooms := make([]string, 0, len(h.rooms))
	var chatusers []string
	for room, users := range h.rooms {
		chatrooms = append(chatrooms, "\""+room+"\":"+strconv.Itoa(len(h.rooms[room])))
		for user := range users {
			var u Client = *user
			chatusers = append(chatusers, u.id)
		}
	}

	data, _ := json.Marshal(&SysMsg{Text: "", RoomInfo: "{" + strings.Join(chatrooms, ",") + "}", UserInfo: strings.Join(chatusers, ",")})
	message, _ := json.Marshal(&Message{Sender: "SYS", RoomId: "", Type: "A", Content: string(data), Time: getUTCTime()})

	// /*發送系統資訊至聊天室*/
	h.sys(message)
	return message
}

/*全頻道廣播*/
func (h *Hub) sys(message []byte) {
	for _, conns := range h.rooms {
		for con := range conns {
			select {
			case con.send <- message:

			default:
				close(con.send)
			}
		}
	}
}

/*這裡可以寫一個讀取終端機輸入然後作為公告發送的功能*/
/*這裡要有一個固定拿redis快取資料的func*/
