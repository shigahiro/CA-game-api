package main

import (
	"net/http"

	"github.com/shigahiro/CA-game-api/handler"

	_ "github.com/go-sql-driver/mysql"
)

func main() {

	http.HandleFunc("/user/create", handler.User_data_create)
	http.HandleFunc("/user/get", handler.User_data_get)
	http.HandleFunc("/user/update", handler.User_data_update)
	http.HandleFunc("/gacha/draw", handler.gachadraw)
	http.HandleFunc("/character/list", handler.character_list)

	http.ListenAndServe(":8080", nil)
}
