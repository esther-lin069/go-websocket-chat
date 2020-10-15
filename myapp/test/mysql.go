package main

import (
	"database/sql"

	//"time"

	_ "github.com/go-sql-driver/mysql"
)

func main() {
	db, err := sql.Open("mysql", "root:root@tcp(mysql)/test?charset=utf8")
	defer db.Close()

	//插入資料
	stmt, err := db.Prepare("INSERT userinfo SET username=?,department=?,created=?")
	checkErr(err)

	_, err = stmt.Exec("astaxie", "研發部門", "2012-12-09")
	checkErr(err)
}

func checkErr(err error) {
	if err != nil {
		panic(err)
	}
}
