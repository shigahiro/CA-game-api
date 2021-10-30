package main

import (
	"database/sql"
	"log"

	_ "github.com/go-sql-driver/mysql"
)

func db_open() (db *sql.DB) {
	db, err := sql.Open("mysql", "root:password@tcp(godockerDB)/sample")
	if err != nil {
		warn.Println("データベース接続失敗")
		log.Fatal(err)
	}
	info.Println("データベース接続成功")
	return db
}
