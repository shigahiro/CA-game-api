package main

import (
	"database/sql"
	"log"
	"net/http"

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

func (*UserHandler) ServeHTTP(w http.ResponseWriter, req *http.Request) {

	db := db_open()
	defer db.Close()

	switch {
	case req.URL.Path == "/user/create" && req.Method == "POST":
		info.Println("ユーザー情報作成ルーティング成功")
		user_data_insert(db, w, req)
	case req.URL.Path == "/user/get" && req.Method == "GET":
		info.Println("ユーザー情報取得ルーティング成功")
		user_data_get(db, w, req)
	case req.URL.Path == "/user/update" && req.Method == "PUT":
		info.Println("ユーザー情報更新ルーティング成功")
		user_data_update(db, w, req)
	case req.URL.Path == "/gacha/draw" && req.Method == "POST":
		info.Println("ガチャ実行ルーティング成功")
		gachadraw(db, w, req)
	case req.URL.Path == "/character/list" && req.Method == "GET":
		info.Println("ユーザ所持キャラ一覧取得ルーティング成功")
		character_list_get(db, w, req)
	default:
		warn.Println("URLまたはリクエストが間違っています")
	}
}
