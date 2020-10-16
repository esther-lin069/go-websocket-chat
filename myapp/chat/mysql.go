package main


import(
	"fmt"

	"database/sql"
	_ "github.com/go-sql-driver/mysql"

)

const(
	host      = "mysql"
	database  = "go_chat"
	user      = "root"
	password  = "root"
)

var db *sql.DB

func checkErr(err error) {
	if err != nil {
		panic(err)
	}
}

func InitDB() *sql.DB{
	var connString = fmt.Sprintf("%s:%s@tcp(%s)/%s?&charset=utf8mb4&collation=utf8mb4_unicode_ci", user, password, host, database)
	db, err := sql.Open("mysql", connString)
	checkErr(err)

	return db
}

func CheckUser(username string){
	//查詢其是否存在
	rows, err := db.Query("SELECT username FROM users WHERE username = ?", username)
	checkErr(err)

	if rows.Next(){
		//exists
	} else{
		//插入資料
		stmt, err := db.Prepare("INSERT users SET username=?,create_at=?")
		checkErr(err)

		_, err = stmt.Exec(username, GetUTCTime())
		checkErr(err)
	}
	
}



