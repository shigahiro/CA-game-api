package main

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"math/rand"
	"net/http"
	"time"

	_ "github.com/go-sql-driver/mysql"
)

type UserHandler struct{}

type User struct {
	ID   int    `json:"id"`
	Name string `json:"name"`
}

type Token struct {
	Token string `json:"token"`
}

func db_open() (db *sql.DB) {
	db, err := sql.Open("mysql", "root:password@tcp(godockerDB)/sample")
	if err != nil {
		log.Fatal(err)
	}
	return db
}

func checkErr(err error) {
	if err != nil {
		panic(err)
	}
}

func RespondWithError(w http.ResponseWriter, code int, msg string) {
	RespondWithJSON(w, code, map[string]string{"error": msg})
}

func RespondWithJSON(w http.ResponseWriter, code int, payload interface{}) {
	response, _ := json.Marshal(payload)
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(code)
	w.Write(response)
}

func RandomString() string {
	var letter = []rune("abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789")

	b := make([]rune, 20)
	for i := range b {
		b[i] = letter[rand.Intn(len(letter))]
	}
	return string(b)
}

func main() {
	handler := &UserHandler{}
	http.Handle("/user/", handler)
	http.ListenAndServe(":8080", nil)

}

func (*UserHandler) ServeHTTP(w http.ResponseWriter, req *http.Request) {

	db := db_open()
	defer db.Close()

	switch {
	case req.URL.Path == "/user/create" && req.Method == "POST":
		user_data_insert(db, w, req)
	case req.URL.Path == "/user/get" && req.Method == "GET":
		user_data_get(db, w, req)
	case req.URL.Path == "/user/update" && req.Method == "POST":
		user_data_update(db, w, req)
	}
}

func user_data_insert(db *sql.DB, w http.ResponseWriter, req *http.Request) {

	var user User

	body, err := ioutil.ReadAll(req.Body)
	defer req.Body.Close()
	if err != nil {
		RespondWithError(w, http.StatusBadRequest, "Invalid request")
		return
	}

	// 読み込んだJSONを構造体に変換
	if err := json.Unmarshal(body, &user); err != nil {
		RespondWithError(w, http.StatusBadRequest, "JSON Unmarshaling failed .")
		return
	}

	token := RandomString()

	stmt, err := db.Prepare("INSERT users SET name=?")
	checkErr(err)
	res, err := stmt.Exec(user.Name)
	checkErr(err)
	id, err := res.LastInsertId()
	checkErr(err)

	stmt, err = db.Prepare("INSERT authentication SET token=?, user_id=?, issued_at=?")
	checkErr(err)
	t := time.Now()
	const layout = "2006-01-02 15:04:05"
	t.Format(layout)
	res, err = stmt.Exec(token, id, t)
	checkErr(err)

	var jsontoken Token
	err = db.QueryRow("SELECT token FROM authentication WHERE user_id = ?", id).Scan(&jsontoken.Token)
	if err != nil {
		log.Fatal(err)
	}

	w.WriteHeader(http.StatusOK)

	json.NewEncoder(w).Encode(jsontoken)
}

func user_data_get(db *sql.DB, w http.ResponseWriter, req *http.Request) {

	type Username struct {
		Name string `json:"name"`
	}
	var username Username
	var id int
	reqtoken := req.Header.Get("x-token")

	err := db.QueryRow("SELECT user_id FROM authentication WHERE token = ?", reqtoken).Scan(&id)
	if err != nil {
		log.Fatal(err)
	}
	err = db.QueryRow("SELECT name FROM users WHERE id = ?", id).Scan(&username.Name)
	if err != nil {
		log.Fatal(err)
	}

	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(username)
}

func user_data_update(db *sql.DB, w http.ResponseWriter, req *http.Request) {

	var user User

	body, err := ioutil.ReadAll(req.Body)
	defer req.Body.Close()
	if err != nil {
		RespondWithError(w, http.StatusBadRequest, "Invalid request")
		return
	}

	// 読み込んだJSONを構造体に変換
	if err := json.Unmarshal(body, &user); err != nil {
		RespondWithError(w, http.StatusBadRequest, "JSON Unmarshaling failed .")
		return
	}

	w.WriteHeader(http.StatusCreated)

	json.NewEncoder(w).Encode(user)
	fmt.Println(user)
	stmt, err := db.Prepare("update users SET name=?  where id=?")
	checkErr(err)
	res, err := stmt.Exec(user.Name, 1)
	checkErr(err)
	affect, err := res.RowsAffected()
	checkErr(err)
	fmt.Println(affect)
}
