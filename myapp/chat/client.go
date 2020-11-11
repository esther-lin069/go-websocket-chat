package main

import (
	"fmt"
	"log"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/go-redis/redis"

	"github.com/gorilla/websocket"
)

const (
	// Time allowed to write a message to the peer.
	writeWait = 10 * time.Second

	// Time allowed to read the next pong message from the peer.
	pongWait = 60 * time.Second

	// Send pings to peer with this period. Must be less than pongWait.
	pingPeriod = (pongWait * 9) / 10

	// Maximum message size allowed from peer.
	maxMessageSize = 512
)

// var (
// 	newline = []byte{'\n'}
// 	space   = []byte{' '}
// )

var upgrader = websocket.Upgrader{
	ReadBufferSize:  1024,
	WriteBufferSize: 1024,
}

//讓redis存的長度保持在 zrange 筆
var zrange int64 = 100

/*
N/normal:普通聊天室訊息
A/all:全域廣播訊息
H/hint:系統提示
I/info:系統資訊
P/private:私訊
WP/私訊通知
*/
type Message struct {
	Sender    string `json:"sender"`
	Recipient string `json:"recipient"`
	RoomId    string `json:"roomId"`
	Type      string `json:"type"`
	Content   string `json:"content"`
	Time      int64  `json:"time"`
}

type RedisMsg struct {
	User  string //client@roomId
	Id    float64
	Value []byte
}

// Client is a middleman between the websocket connection and the hub.
type Client struct {
	//id string
	id string

	roomId string

	roomType string

	hub *Hub

	// The websocket connection.
	conn *websocket.Conn

	redis_conn *redis.Client

	// Buffered channel of outbound messages.
	send chan []byte
}

func GetUTCTime() string {
	tn := time.Now()
	local, err := time.LoadLocation("UTC")
	if err != nil {
		fmt.Println(err)
	}
	t := tn.In(local)
	formatted := fmt.Sprintf("%d-%02d-%02d %02d:%02d:%02d",
		t.Year(), t.Month(), t.Day(),
		t.Hour(), t.Minute(), t.Second())
	return formatted
}

/*讀取從ws來的訊息*/
func (c *Client) readPump() {
	defer func() {
		c.hub.unregister <- c
		c.conn.Close()
	}()
	c.conn.SetReadLimit(maxMessageSize)
	c.conn.SetReadDeadline(time.Now().Add(pongWait))
	c.conn.SetPongHandler(func(string) error { c.conn.SetReadDeadline(time.Now().Add(pongWait)); return nil })
	for {
		_, message, err := c.conn.ReadMessage()
		if err != nil {
			if websocket.IsUnexpectedCloseError(err, websocket.CloseGoingAway, websocket.CloseAbnormalClosure) {
				log.Printf("error: %v", err)
			}
			break
		}

		/*存入redis*/
		m := RedisMsg{c.roomId, float64(time.Now().UnixNano()), message}
		c.ZsetMessage(m) //以毫秒作為key

		/*存入Mysql*/
		PutMsgSingle(message)

		c.hub.broadcast <- message
	}
}

/*將訊息寫入ws連線*/
func (c *Client) writePump() {
	ticker := time.NewTicker(pingPeriod)
	defer func() {
		ticker.Stop()
		c.conn.Close()
	}()
	for {
		select {
		case message, ok := <-c.send:
			c.conn.SetWriteDeadline(time.Now().Add(writeWait))
			if !ok {
				// The hub closed the channel.
				c.conn.WriteMessage(websocket.CloseMessage, []byte{})
				return
			}

			w, err := c.conn.NextWriter(websocket.TextMessage)
			if err != nil {
				return
			}
			w.Write(message)

			if err := w.Close(); err != nil {
				return
			}
		case <-ticker.C:
			c.conn.SetWriteDeadline(time.Now().Add(writeWait))
			if err := c.conn.WriteMessage(websocket.PingMessage, nil); err != nil {
				return
			}
		}
	}
}

/*serveWs handles websocket requests from the peer.*/
func serveWs(hub *Hub, ctx *gin.Context) {
	conn, err := upgrader.Upgrade(ctx.Writer, ctx.Request, nil)
	if err != nil {
		log.Println(err)
		return
	}
	username := ctx.Query("user")
	room := ctx.Param("roomId")
	private := ctx.Query("private")

	var roomType string
	if private == "true" {
		roomType = "private"
	} else {
		roomType = "normal"
	}

	fmt.Println("user:" + username + "/ room:" + room + " .registered type:" + roomType)

	client := &Client{id: username, roomId: room, roomType: roomType, hub: hub, redis_conn: GetRedisClient(), conn: conn, send: make(chan []byte, 256)}
	client.hub.loadmsg <- client
	client.hub.register <- client

	// Allow collection of memory referenced by the caller by doing all work in new goroutines
	go client.writePump()
	go client.readPump()
}
