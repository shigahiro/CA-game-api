package main

import (
	cryptorand "crypto/rand"
	"encoding/json"
	"errors"
	"io/ioutil"
	"net/http"
	"time"

	"github.com/shigahiro/CA-game-api/model"
)

func checkErr(err error, errstring string) {
	if err != nil {
		model.Warn.Println(errstring)
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
		model.Warn.Println("ボディを読み取るのに失敗しました")
		return err
	}

	// 読み込んだJSONを構造体に変換
	if err := json.Unmarshal(body, i); err != nil {
		RespondWithError(w, http.StatusBadRequest, "JSON Unmarshaling failed .")
		model.Warn.Println("JSONを構造体に変換できませんでした")
		return err
	}
	return err
}

func User_data_create(w http.ResponseWriter, req *http.Request) {
	model.Info.Println("ユーザー情報作成ルーティング成功")

	var user model.User
	var i interface{}
	i = &user

	if err := unmarshalingjson(i, w, req); err != nil {
		return
	}

	token, err := RandomString()
	checkErr(err, "トークン生成失敗")
	model.Info.Println("トークン生成成功")

	db := db_open()
	defer db.Close()
	stmt, err := db.Prepare("INSERT users SET name=?")
	checkErr(err, "Stmtオブジェクト生成失敗")
	model.Info.Println("Stmtオブジェクト生成成功")
	res, err := stmt.Exec(user.Name)
	checkErr(err, "ユーザ情報の挿入失敗")
	model.Info.Println("ユーザ情報の挿入成功")
	id, err := res.LastInsertId()
	checkErr(err, "user_id取得失敗")
	checkErr(err, "user_id取得成功")

	stmt, err = db.Prepare("INSERT authentication SET token=?, user_id=?, issued_at=?")
	checkErr(err, "Stmtオブジェクト生成失敗")
	model.Info.Println("Stmtオブジェクト生成成功")
	t := time.Now()
	const layout = "2006-01-02 15:04:05"
	t.Format(layout)
	res, err = stmt.Exec(token, id, t)
	checkErr(err, "認証情報の挿入失敗")
	model.Info.Println("認証情報の挿入成功")

	var jsontoken model.Token
	jsontoken.Token = token

	w.WriteHeader(http.StatusOK)

	json.NewEncoder(w).Encode(jsontoken)
}

func User_data_update(w http.ResponseWriter, req *http.Request) {
	model.Info.Println("ユーザー情報更新ルーティング成功")

	var user model.User
	var i interface{}
	i = &user
	if err := unmarshalingjson(i, w, req); err != nil {
		return
	}

	reqtoken := req.Header.Get("x-token")
	if reqtoken == "" {
		model.Warn.Println("トークンを取得できません")
		return
	}
	model.Info.Println("トークン取得成功")

	db := db_open()
	defer db.Close()
	err := db.QueryRow("SELECT user_id FROM authentication WHERE token = ?", reqtoken).Scan(&user.ID)
	checkErr(err, "トークンがありません")
	model.Info.Println("トークンが一致しました")

	stmt, err := db.Prepare("update users SET name=? where id=?")
	checkErr(err, "Stmtオブジェクト生成失敗")
	model.Info.Println("Stmtオブジェクト生成成功")
	_, err = stmt.Exec(user.Name, user.ID)
	checkErr(err, "ユーザテーブルの更新失敗")
	model.Info.Println("ユーザテーブルの更新成功")

	stmt, err = db.Prepare("UPDATE authentication SET issued_at=? where user_id=?")
	checkErr(err, "Stmtオブジェクト生成失敗")
	model.Info.Println("Stmtオブジェクト生成成功")
	t := time.Now()
	const layout = "2006-01-02 15:04:05"
	t.Format(layout)
	_, err = stmt.Exec(t, user.ID)
	checkErr(err, "ユーザテーブルの更新失敗")
	model.Info.Println("ユーザテーブルの更新成功")

	w.WriteHeader(http.StatusOK)

	json.NewEncoder(w).Encode(user)
}

func User_data_get(w http.ResponseWriter, req *http.Request) {
	model.Info.Println("ユーザー情報取得ルーティング成功")

	type Username struct {
		Name string `json:"name"`
	}
	var username Username
	var id int

	reqtoken := req.Header.Get("x-token")
	if reqtoken == "" {
		model.Warn.Println("トークンを取得できません")
		return
	}
	model.Info.Println("トークン取得成功")

	db := db_open()
	defer db.Close()
	err := db.QueryRow("SELECT user_id FROM authentication WHERE token = ?", reqtoken).Scan(&id)
	checkErr(err, "トークンがありません")
	model.Info.Println("トークンが一致しました")

	err = db.QueryRow("SELECT name FROM users WHERE id = ?", id).Scan(&username.Name)
	checkErr(err, "ユーザ情報を取得に失敗しました")
	model.Info.Println("ユーザ情報を取得しました")

	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(username)
}
