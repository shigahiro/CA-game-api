package main

import (
	"database/sql"
	"fmt"
	"log"
	"net/http"

	_ "github.com/go-sql-driver/mysql"
)

type helloHandler struct{}

func (*helloHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	db, err := sql.Open("mysql", "root:password@tcp(godockerDB)/sample")
	if err != nil {
		log.Fatal(err)
	}
	defer db.Close()
	switch {
	case r.URL.Path == "/user/create" && r.Method == "POST":
		fmt.Println("create")
	case r.URL.Path == "/user/get" && r.Method == "GET":
		fmt.Println("update")
	case r.URL.Path == "/user/update" && r.Method == "POST":
		fmt.Println("update")
	}
}

func main() {

	handler := &helloHandler{}
	http.Handle("/user/", handler)

	http.ListenAndServe(":8080", nil)

}
