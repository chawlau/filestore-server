package mysql

import (
	"database/sql"
	"fmt"
	"os"

	_ "github.com/go-sql-driver/mysql"
)

var db *sql.DB

func init() {
	tmpDB, err := sql.Open("mysql", "root:L19880901c-@tcp(192.168.31.186:3306)/fileserver")
	if err != nil {
		fmt.Println("connect db failed err ", err.Error())
	}
	db = tmpDB

	db.SetMaxOpenConns(1000)
	err = db.Ping()

	if err != nil {
		fmt.Println("Failed to connect to mysql err ", err.Error())
		os.Exit(1)
	}
}

func DBConn() *sql.DB {
	return db
}
