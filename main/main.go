package main

import (
	"log"
	"net/http"
	"os"

	_ "github.com/go-sql-driver/mysql"
)

var (
	warn = log.New(os.Stderr, "[WARN] ", log.Ldate|log.Ltime|log.Lshortfile)
	info = log.New(os.Stdout, "[INFO] ", log.Ldate|log.Ltime|log.Lshortfile)
)

func main() {

	http.HandleFunc("/user/create", user_data_create)
	http.HandleFunc("/user/get", user_data_get)
	http.HandleFunc("/user/update", user_data_update)
	http.HandleFunc("/gacha/draw", gachadraw)
	http.HandleFunc("/character/list", character_list)

	http.ListenAndServe(":8080", nil)
}
