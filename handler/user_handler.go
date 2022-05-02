package handler

import (
	cryptorand "crypto/rand"
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"net/http"
	"time"

	"github.com/shigahiro/CA-game-api/model"
)

func RespondWithError(w http.ResponseWriter, code int, msg string) {
	RespondWithJSON(w, code, map[string]string{"error": msg})
}

func RespondWithJSON(w http.ResponseWriter, code int, payload interface{}) {
	response, _ := json.Marshal(payload)
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(code)
	w.Write(response)
}

func randomString() (string, error) {
	const letters = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789"

	// 乱数を生成
	b := make([]byte, 20)
	if _, err := cryptorand.Read(b); err != nil {
		return "", errors.New("unexpected error")
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
	var i interface{} = &user

	if err := unmarshalingjson(i, w, req); err != nil {
		return
	}

	fmt.Println(i)
	token, err := randomString()
	if err != nil {
		model.Warn.Println("トークン生成失敗")
		return
	}
	model.Info.Println("トークン生成成功")

	db := db_open()
	defer db.Close()

	t := time.Now()
	const layout = "2006-01-02 15:04:05"
	t.Format(layout)

	stmt, err := db.Prepare("INSERT users SET name=?, token=?, issued_at=?")
	if err != nil {
		model.Warn.Println("Stmtオブジェクト生成失敗")
		return
	}
	model.Info.Println("Stmtオブジェクト生成成功")
	_, err = stmt.Exec(user.Name, token, t)
	if err != nil {
		model.Warn.Println("認証情報の挿入失敗")
		return
	}
	model.Info.Println("認証情報の挿入成功")

	var jsontoken model.Token
	jsontoken.Token = token

	w.WriteHeader(http.StatusOK)

	json.NewEncoder(w).Encode(jsontoken)
}

func User_data_update(w http.ResponseWriter, req *http.Request) {
	model.Info.Println("ユーザー情報更新ルーティング成功")

	var user model.User
	var i interface{} = &user
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
	err := db.QueryRow("SELECT id FROM users WHERE token = ?", reqtoken).Scan(&user.ID)
	if err != nil {
		model.Warn.Println("トークンがありません")
		return
	}
	model.Info.Println("トークンが一致しました")

	stmt, err := db.Prepare("update users SET name=?, issued_at=? where id=?")
	if err != nil {
		model.Warn.Println("Stmtオブジェクト生成失敗")
		return
	}
	model.Info.Println("Stmtオブジェクト生成成功")

	t := time.Now()
	const layout = "2006-01-02 15:04:05"
	t.Format(layout)

	_, err = stmt.Exec(user.Name, t, user.ID)
	if err != nil {
		model.Warn.Println("ユーザテーブルの更新失敗")
		return
	}
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
	err := db.QueryRow("SELECT id FROM users WHERE token = ?", reqtoken).Scan(&id)
	if err != nil {
		model.Warn.Println("トークンがありません")
		return
	}
	model.Info.Println("トークンが一致しました")

	err = db.QueryRow("SELECT name FROM users WHERE id = ?", id).Scan(&username.Name)
	if err != nil {
		model.Warn.Println("ユーザ情報の取得に失敗しました")
		return
	}
	model.Info.Println("ユーザ情報を取得しました")

	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(username)
}
