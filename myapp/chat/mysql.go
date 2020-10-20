package main

import (
	"fmt"

	"database/sql"

	_ "github.com/go-sql-driver/mysql"
)

const (
	host     = "mysql"
	database = "go_chat"
	user     = "root"
	password = "root"
)

var db *sql.DB

func checkErr(err error) {
	if err != nil {
		panic(err)
	}
}

func InitDB() *sql.DB {
	var connString = fmt.Sprintf("%s:%s@tcp(%s)/%s?&charset=utf8mb4&collation=utf8mb4_unicode_ci", user, password, host, database)
	db, err := sql.Open("mysql", connString)
	checkErr(err)

	return db
}

func CheckUser(username string) {
	//查詢其是否存在
	rows, err := db.Query("SELECT username FROM users WHERE username = ?", username)
	checkErr(err)

	if rows.Next() {
		//exists
	} else {
		//插入資料
		stmt, err := db.Prepare("INSERT users SET username=?,create_at=?")
		checkErr(err)

		_, err = stmt.Exec(username, GetUTCTime())
		checkErr(err)
	}

}

func MakeRoom(roomId string) {
	//查詢其是否存在
	rows, err := db.Query("SELECT roomId FROM rooms WHERE roomId = ?", roomId)
	checkErr(err)

	if rows.Next() {
		//exists
	} else {
		//插入資料
		stmt, err := db.Prepare("INSERT rooms SET roomId=?,create_at=?")
		checkErr(err)

		_, err = stmt.Exec(roomId, GetUTCTime())
		checkErr(err)
	}
}

func MakeUser_RoomCheck(username string, roomId string) {

	rows, err := db.Query("SELECT * FROM `user-room` WHERE user_id = ? AND room_id = ?", username, roomId)
	checkErr(err)
	defer rows.Close()

	if rows.Next() {
		//exists
	} else {
		//插入資料
		stmt, err := db.Prepare("INSERT `user-room` SET user_id=?, room_id = ?, create_at=?")
		checkErr(err)

		_, err = stmt.Exec(username, roomId, GetUTCTime())
		checkErr(err)
	}

}

func GetRoomList(username string) []string {
	var roomList []string
	var room string
	rows, err := db.Query("SELECT room_id FROM `user-room` WHERE user_id = ?", username)
	checkErr(err)

	for rows.Next() {
		err := rows.Scan(&room)
		checkErr(err)

		roomList = append(roomList, room)
	}

	return roomList
}

func GetUserList() []string {
	var userList []string
	var user string
	rows, err := db.Query("SELECT username FROM `users`")
	checkErr(err)

	for rows.Next() {
		err := rows.Scan(&user)
		checkErr(err)

		userList = append(userList, user)
	}

	return userList
}

func DelRoom(roomId string) {
	fmt.Print("delete room:")
	fmt.Println(roomId)
	_, err := db.Exec("DELETE FROM rooms WHERE roomId = ?", roomId)
	checkErr(err)
}

func LeaveRoom(roomId string, user string) {
	_, err := db.Exec("DELETE FROM `user-room` WHERE room_id = ? AND user_id = ?", roomId, user)
	checkErr(err)
}
