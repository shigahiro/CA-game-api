package main

import (
	"database/sql"
	"fmt"
	"log"

	_ "github.com/go-sql-driver/mysql"
)

func main() {
	db, err := sql.Open("mysql", "root:password@tcp(godockerDB)/sample")
	if err != nil {
		log.Fatal(err)
	}
	defer db.Close()
	id := 1
	var name string

	if err := db.QueryRow("SELECT name FROM users").Scan(&name); err != nil {
		log.Fatal(err)
	}

	fmt.Println(id, name)
}
