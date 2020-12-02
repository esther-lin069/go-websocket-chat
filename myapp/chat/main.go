package main

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
)

type MakeRoomInfo struct {
	UserName string `json:"user"`
	RoomName string `json:"roomName"`
	With     string `json:"with"`
}

func serveHome(ctx *gin.Context) {
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
	ctx.HTML(http.StatusOK, "index.html", gin.H{
		"ipAddress": ctx.ClientIP(),
	})
}

func login(ctx *gin.Context) {
	username := strings.Trim(ctx.Request.FormValue("username"), " ")
	if username == "" {
		return
	}
	CheckUser(username)
	ctx.Redirect(http.StatusMovedPermanently, "/chat/main/?user="+username+"&private=false")

}

func makePrivateRoom(ctx *gin.Context) {
	data, _ := ctx.GetRawData()
	// 解析post來的資料
	var mpi MakeRoomInfo
	err := json.Unmarshal(data, &mpi)
	if err != nil {
		fmt.Println(err)
		return
	}
	if mpi.RoomName == "" || mpi.UserName == "" {
		return
	}
	// 儲存私訊房間紀錄
	MakePrivateRoom(mpi.RoomName)
	// 已讀狀態
	HsetForPrivate(mpi.UserName, mpi.With, "1")
	// 進入私聊房間
	ctx.Redirect(http.StatusFound, "/chat/"+mpi.RoomName+"?user="+mpi.UserName+"&private=true")

}

func makeNormalRoom(ctx *gin.Context) {
	data, _ := ctx.GetRawData()
	// 解析post來的資料
	var mri MakeRoomInfo
	err := json.Unmarshal(data, &mri)
	if err != nil {
		fmt.Println(err)
		return
	}
	if mri.RoomName == "" || mri.UserName == "" {
		return
	}
	// 建立房間
	MakeRoom(mri.RoomName)
	// 將使用者寫入該房間
	MakeUser_RoomCheck(mri.UserName, mri.RoomName)
	// 重新導向至該房間
	ctx.Redirect(http.StatusFound, "/chat/"+mri.RoomName+"?user="+mri.UserName+"&private=false")
}

func askRoomList(ctx *gin.Context) {
	username := ctx.Query("user")
	list := GetRoomList(username)
	ctx.JSON(200, gin.H{
		"rooms": strings.Join(list, ","),
	})

}

func askFakeList(ctx *gin.Context) {
	list := []string{
		"Room1",
		"Room2",
		"Room3",
		"Room4",
	}
	ctx.JSON(200, gin.H{
		"rooms": strings.Join(list, ","),
	})
}

func askUserList(hub *Hub, ctx *gin.Context) {
	list := GetUserList()
	//online_list := hub.makeInfo("get")
	ctx.JSON(200, gin.H{
		"users": strings.Join(list, ","),
	})
}

//刪除房間
func doDelRoom(ctx *gin.Context) {
	roomId := ctx.Param("roomId")
	DelRoom(roomId)
	ctx.Redirect(http.StatusFound, "/chat/main/?user="+ctx.Query("user")+"&private=false") //進入聊天室
}

//退出房間
func doLeaveRoom(ctx *gin.Context) {
	roomId := ctx.Param("roomId")
	user := ctx.Query("user")
	LeaveRoom(roomId, user)
	ctx.Redirect(http.StatusFound, "/chat/main/?user="+user+"&private=false") //進入聊天室
}

func readStatus(ctx *gin.Context) {
	user := ctx.Param("user")
	result := GetHashForPrivate(user)
	json, err := json.Marshal(result)
	if err != nil {
		fmt.Println(err)
	}
	ctx.JSON(200, gin.H{
		"readstatus": string(json),
	})
}

func main() {
	hub := newHub()
	db = InitDB()
	defer db.Close()

	go hub.run()
	go hub.sysTicker()
	// ROUTER
	router := gin.Default()
	router.Delims("{[{", "}]}") //自定義模板隔符避免與Vue衝突
	// router.LoadHTMLFiles("public/home.html", "public/login.html")
	// router.Static("/asset", "./asset")

	router.LoadHTMLFiles("dist/index.html", "public/login.html", "chat_window/chat_index.html", "public/test.html")
	router.Static("/assets", "./dist/assets")
	router.Static("/chat_assets", "./chat_window/assets")

	router.POST("/login", login)

	router.GET("/roomlist", askRoomList)
	router.GET("/fakelist", askFakeList)
	router.GET("/userlist", func(ctx *gin.Context) {
		askUserList(hub, ctx)
	})

	router.GET("/delete/:roomId", doDelRoom)
	router.GET("/leave/:roomId", doLeaveRoom)
	router.GET("/readstatus/:user", readStatus)
	router.GET("/chat/:roomId", serveHome)

	router.POST("/privateroom", makePrivateRoom)
	router.POST("/normalroom", makeNormalRoom)

	router.GET("/login", func(ctx *gin.Context) {
		ctx.HTML(http.StatusOK, "login.html", nil)
	})
	router.GET("/ws/chat/:roomId", func(ctx *gin.Context) {
		serveWs(hub, ctx)
	})

	// router.GET("/info", func(ctx *gin.Context) {
	// 	data := hub.makeInfo("get")
	// 	ctx.JSON(200, gin.H{
	// 		"Rooms": data[0],
	// 		"Users": data[1],
	// 	})
	// })
	router.GET("/dist", func(ctx *gin.Context) {
		ctx.HTML(http.StatusOK, "chat_index.html", nil)
	})

	router.GET("/test", func(ctx *gin.Context) {
		ctx.HTML(http.StatusOK, "test.html", nil)
	})
	router.GET("/", serveHome)

	router.Run(":8080")

}
