package main

import (
	"crypto/rand"
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"log"
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

// type Character {
// 	UsercharacterID string

// }

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

func RandomString() (string, error) {
	const letters = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789"

	// 乱数を生成
	b := make([]byte, 20)
	if _, err := rand.Read(b); err != nil {
		return "", errors.New("unexpected error...")
	}

	// letters からランダムに取り出して文字列を生成
	var result string
	for _, v := range b {
		// index が letters の長さに収まるように調整
		result += string(letters[int(v)%len(letters)])
	}
	return result, nil
}

func main() {
	handler := &UserHandler{}
	http.Handle("/", handler)
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
	case req.URL.Path == "/user/update" && req.Method == "PUT":
		user_data_update(db, w, req)
	case req.URL.Path == "/gacha/draw" && req.Method == "POST":
		fmt.Println("gachadraw")
	case req.URL.Path == "/character/list" && req.Method == "GET":
		fmt.Println("characterlist")
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

	token, _ := RandomString()

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
	checkErr(err)

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
	checkErr(err)
	err = db.QueryRow("SELECT name FROM users WHERE id = ?", id).Scan(&username.Name)
	checkErr(err)

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

	reqtoken := req.Header.Get("x-token")

	err = db.QueryRow("SELECT user_id FROM authentication WHERE token = ?", reqtoken).Scan(&user.ID)
	checkErr(err)

	stmt, err := db.Prepare("update users SET name=? where id=?")
	checkErr(err)
	res, err := stmt.Exec(user.Name, user.ID)
	checkErr(err)
	affect, err := res.RowsAffected()
	checkErr(err)
	fmt.Println(affect)

	stmt, err = db.Prepare("UPDATE authentication SET issued_at=? where user_id=?")
	checkErr(err)
	t := time.Now()
	const layout = "2006-01-02 15:04:05"
	t.Format(layout)
	res, err = stmt.Exec(t, user.ID)
	checkErr(err)

	w.WriteHeader(http.StatusOK)

	json.NewEncoder(w).Encode(user)
}

// func gachadraw {
// 	math.rand
// }
