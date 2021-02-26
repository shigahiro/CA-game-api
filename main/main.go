package main

import (
	"net/http"

	_ "github.com/go-sql-driver/mysql"
)

func main() {
	handler := &UserHandler{}
	http.Handle("/", handler)
	http.ListenAndServe(":8080", nil)
}
