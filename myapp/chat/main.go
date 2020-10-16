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
		pusers := strings.Split(ctx.Query("room"), "-")
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
	ctx.Redirect(http.StatusMovedPermanently, "/?user="+username+"&room=main&private=false")

}

func makePrivateRoom(ctx *gin.Context) {
	username := ctx.Request.FormValue("user")
	roomName := ctx.Request.FormValue("roomName")
	ctx.Redirect(http.StatusMovedPermanently, "/?user="+username+"&room="+roomName+"&private=ture") //進入聊天室

}

func main() {
	hub := newHub()
	go hub.run()

	// ROUTER
	router := gin.Default()
	router.LoadHTMLGlob("public/*")

	router.GET("/login", func(ctx *gin.Context) {
		ctx.HTML(http.StatusOK, "login.html", nil)
	})

	router.POST("/login", login)

	router.GET("/", serveHome)

	router.POST("/privateroom", makePrivateRoom)

	router.GET("/ws", func(ctx *gin.Context) {
		serveWs(hub, ctx)
	})

	router.GET("/info", func(ctx *gin.Context) {
		data := hub.makeInfo()
		ctx.JSON(200, gin.H{
			"Rooms": data[0],
			"Users": data[1],
		})
	})

	router.Run(":8080")

}
