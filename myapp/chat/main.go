package main

import (
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
)

func serveHome(ctx *gin.Context) {
	//log.Println(r.URL)
	if ctx.Query("user") == "" {
		ctx.Redirect(http.StatusMovedPermanently, "/login") //在這裡加私訊驗證
	}
	if ctx.Query("private") == "true" {
		user := ctx.Query("user")
		pusers := strings.Split(ctx.Param("roomId"), "-")
		if len(pusers) != 2 {
			ctx.JSON(400, gin.H{
				"error": "missing RoomID",
			})
			return
		}
		if !(user == pusers[0] || user == pusers[1]) {
			ctx.JSON(400, gin.H{
				"error": "wrong Private RoomID",
			})
			return
		}

	}
	ctx.HTML(http.StatusOK, "home.html", nil)
}

func login(ctx *gin.Context) {
	username := strings.Trim(ctx.Request.FormValue("username"), " ")
	CheckUser(username)
	ctx.Redirect(http.StatusMovedPermanently, "/chat/main/?user="+username+"&private=false")

}

func makePrivateRoom(ctx *gin.Context) {
	username := ctx.Request.FormValue("user")
	roomName := ctx.Request.FormValue("roomName")
	ctx.Redirect(http.StatusMovedPermanently, "/chat/"+roomName+"?user="+username+"&private=true") //進入聊天室

}

func makeNormalRoom(ctx *gin.Context) {
	username := ctx.Request.FormValue("user")
	roomName := ctx.Request.FormValue("roomName")
	MakeRoom(roomName)
	MakeUser_RoomCheck(username, roomName)
	ctx.Redirect(http.StatusMovedPermanently, "/chat/"+roomName+"?user="+username+"&private=false") //進入聊天室
}

func askRoomList(ctx *gin.Context) {
	username := ctx.Request.FormValue("user")
	list := GetRoomList(username)
	ctx.JSON(200, gin.H{
		"rooms": strings.Join(list, ","),
	})

}

func askUserList(ctx *gin.Context) {
	list := GetUserList()
	ctx.JSON(200, gin.H{
		"users": strings.Join(list, ","),
	})
}

//刪除房間
func doDelRoom(ctx *gin.Context) {
	roomId := ctx.Param("roomId")
	ctx.JSON(200, gin.H{
		"RoomId": roomId,
	})
	DelRoom(roomId)

}

//退出房間

func main() {
	hub := newHub()
	db = InitDB()
	defer db.Close()

	go hub.run()

	// ROUTER
	router := gin.Default()
	router.LoadHTMLGlob("public/*")

	router.GET("/login", func(ctx *gin.Context) {
		ctx.HTML(http.StatusOK, "login.html", nil)
	})

	router.POST("/login", login)

	router.POST("/roomlist", askRoomList)
	router.GET("/userlist", askUserList)

	router.GET("/chat/:roomId", serveHome)
	router.GET("/delete/:roomId", doDelRoom)

	router.POST("/privateroom", makePrivateRoom)
	router.POST("/normalroom", makeNormalRoom)

	router.GET("/ws/chat/:roomId", func(ctx *gin.Context) {
		serveWs(hub, ctx)
	})

	router.GET("/info", func(ctx *gin.Context) {
		data := hub.makeInfo()
		ctx.JSON(200, gin.H{
			"Rooms": data[0],
			"Users": data[1],
		})
	})

	router.GET("/", serveHome)

	router.Run(":8080")

}
