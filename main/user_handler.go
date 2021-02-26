package main

import (
	cryptorand "crypto/rand"
	"database/sql"
	"encoding/json"
	"errors"
	"io/ioutil"
	"net/http"
	"time"
)

func checkErr(err error, errstring string) {
	if err != nil {
		warn.Println(errstring)
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
	if _, err := cryptorand.Read(b); err != nil {
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

func unmarshalingjson(i interface{}, w http.ResponseWriter, req *http.Request) error {

	body, err := ioutil.ReadAll(req.Body)
	defer req.Body.Close()
	if err != nil {
		RespondWithError(w, http.StatusBadRequest, "Invalid request")
		warn.Println("ボディを読み取るのに失敗しました")
		return err
	}

	// 読み込んだJSONを構造体に変換
	if err := json.Unmarshal(body, i); err != nil {
		RespondWithError(w, http.StatusBadRequest, "JSON Unmarshaling failed .")
		warn.Println("JSONを構造体に変換できませんでした")
		return err
	}
	return err
}

func user_data_insert(db *sql.DB, w http.ResponseWriter, req *http.Request) {

	var user User
	var i interface{}
	i = &user

	if err := unmarshalingjson(i, w, req); err != nil {
		return
	}

	token, err := RandomString()
	checkErr(err, "トークン生成失敗")
	info.Println("トークン生成成功")

	stmt, err := db.Prepare("INSERT users SET name=?")
	checkErr(err, "Stmtオブジェクト生成失敗")
	info.Println("Stmtオブジェクト生成成功")
	res, err := stmt.Exec(user.Name)
	checkErr(err, "ユーザ情報の挿入失敗")
	info.Println("ユーザ情報の挿入成功")
	id, err := res.LastInsertId()
	checkErr(err, "user_id取得失敗")
	checkErr(err, "user_id取得成功")

	stmt, err = db.Prepare("INSERT authentication SET token=?, user_id=?, issued_at=?")
	checkErr(err, "Stmtオブジェクト生成失敗")
	info.Println("Stmtオブジェクト生成成功")
	t := time.Now()
	const layout = "2006-01-02 15:04:05"
	t.Format(layout)
	res, err = stmt.Exec(token, id, t)
	checkErr(err, "認証情報の挿入失敗")
	info.Println("認証情報の挿入成功")

	var jsontoken Token
	jsontoken.Token = token

	w.WriteHeader(http.StatusOK)

	json.NewEncoder(w).Encode(jsontoken)
}

func user_data_update(db *sql.DB, w http.ResponseWriter, req *http.Request) {

	var user User
	var i interface{}
	i = &user
	if err := unmarshalingjson(i, w, req); err != nil {
		return
	}

	reqtoken := req.Header.Get("x-token")
	if reqtoken == "" {
		warn.Println("トークンを取得できません")
		return
	}
	info.Println("トークン取得成功")

	err := db.QueryRow("SELECT user_id FROM authentication WHERE token = ?", reqtoken).Scan(&user.ID)
	checkErr(err, "トークンがありません")
	info.Println("トークンが一致しました")

	stmt, err := db.Prepare("update users SET name=? where id=?")
	checkErr(err, "Stmtオブジェクト生成失敗")
	info.Println("Stmtオブジェクト生成成功")
	_, err = stmt.Exec(user.Name, user.ID)
	checkErr(err, "ユーザテーブルの更新失敗")
	info.Println("ユーザテーブルの更新成功")

	stmt, err = db.Prepare("UPDATE authentication SET issued_at=? where user_id=?")
	checkErr(err, "Stmtオブジェクト生成失敗")
	info.Println("Stmtオブジェクト生成成功")
	t := time.Now()
	const layout = "2006-01-02 15:04:05"
	t.Format(layout)
	_, err = stmt.Exec(t, user.ID)
	checkErr(err, "ユーザテーブルの更新失敗")
	info.Println("ユーザテーブルの更新成功")

	w.WriteHeader(http.StatusOK)

	json.NewEncoder(w).Encode(user)
}

func user_data_get(db *sql.DB, w http.ResponseWriter, req *http.Request) {

	type Username struct {
		Name string `json:"name"`
	}
	var username Username
	var id int

	reqtoken := req.Header.Get("x-token")
	if reqtoken == "" {
		warn.Println("トークンを取得できません")
		return
	}
	info.Println("トークン取得成功")

	err := db.QueryRow("SELECT user_id FROM authentication WHERE token = ?", reqtoken).Scan(&id)
	checkErr(err, "トークンがありません")
	info.Println("トークンが一致しました")

	err = db.QueryRow("SELECT name FROM users WHERE id = ?", id).Scan(&username.Name)
	checkErr(err, "ユーザ情報を取得に失敗しました")
	info.Println("ユーザ情報を取得しました")

	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(username)
}
