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
		log.Fatal(err)
	}
	return db
}

func (*UserHandler) ServeHTTP(w http.ResponseWriter, req *http.Request) {

	db := db_open()
	defer db.Close()

	switch {
	case req.URL.Path == "/user/create" && req.Method == "POST":
		user_data_insert(db, w, req)
	case req.URL.Path == "/user/get" && req.Method == "GET":
		user_data_get(db, w, req)
	case req.URL.Path == "/user/update" && req.Method == "PUT":
		user_data_update(db, w, req)
	case req.URL.Path == "/gacha/draw" && req.Method == "POST":
		gachadraw(db, w, req)
	case req.URL.Path == "/character/list" && req.Method == "GET":
		character_list_get(db, w, req)
	}
}
