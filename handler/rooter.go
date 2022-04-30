package main

import (
	"database/sql"
	"log"

	"github.com/shigahiro/CA-game-api/model"

	_ "github.com/go-sql-driver/mysql"
)

func db_open() (db *sql.DB) {
	db, err := sql.Open("mysql", "root:password@tcp(godockerDB)/sample")
	if err != nil {
		model.Warn.Println("データベース接続失敗")
		log.Fatal(err)
	}
	model.Info.Println("データベース接続成功")
	return db
}
