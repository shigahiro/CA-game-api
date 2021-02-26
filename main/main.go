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
	handler := &UserHandler{}
	http.Handle("/", handler)
	http.ListenAndServe(":8080", nil)
}
