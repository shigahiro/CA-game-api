package main

import (
	crypto "crypto/rand"
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"log"
	"math/rand"
	"net/http"
	"time"
)

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
	if _, err := crypto.Read(b); err != nil {
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
		return err
	}

	// 読み込んだJSONを構造体に変換
	if err := json.Unmarshal(body, i); err != nil {
		RespondWithError(w, http.StatusBadRequest, "JSON Unmarshaling failed .")
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

func user_data_update(db *sql.DB, w http.ResponseWriter, req *http.Request) {

	var user User
	var i interface{}
	i = &user
	if err := unmarshalingjson(i, w, req); err != nil {
		return
	}

	reqtoken := req.Header.Get("x-token")

	err := db.QueryRow("SELECT user_id FROM authentication WHERE token = ?", reqtoken).Scan(&user.ID)
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

func character_list_get(db *sql.DB, w http.ResponseWriter, req *http.Request) {
	var id int
	reqtoken := req.Header.Get("x-token")

	err := db.QueryRow("SELECT user_id FROM authentication WHERE token = ?", reqtoken).Scan(&id)
	checkErr(err)

	rows, err := db.Query("SELECT user_id, character_id, character_name FROM possession_characters WHERE user_id = ?", id)
	checkErr(err)

	defer rows.Close()

	var possession_character Possession_character
	var possession_characters Possession_characters
	for rows.Next() {
		if err := rows.Scan(&possession_character.UserID, &possession_character.CharacterID, &possession_character.Name); err != nil {
			log.Fatal(err)
		}
		possession_characters.Characters = append(possession_characters.Characters, possession_character)

	}
	if err := rows.Err(); err != nil {
		log.Fatal(err)
	}

	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(possession_characters)
}

func gachadraw(db *sql.DB, w http.ResponseWriter, req *http.Request) {
	var results Results
	var gacha_times GachaTime

	var i interface{}
	i = &gacha_times
	if err := unmarshalingjson(i, w, req); err != nil {
		return
	}

	// time := gacha_times.Time
	for i := gacha_times.Time; i >= 1; i-- {
		randam := rand.Intn(100)
		switch {
		case 0 <= randam && randam < 4:
			s := "S"
			result := gatya_data_insert(db, w, req, s)
			results.Results = append(results.Results, result)
		case 4 <= randam && randam < 20:
			s := "A"
			result := gatya_data_insert(db, w, req, s)
			results.Results = append(results.Results, result)
		default:
			s := "B"
			result := gatya_data_insert(db, w, req, s)
			results.Results = append(results.Results, result)
		}
	}

	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(results)
}

func gatya_data_insert(db *sql.DB, w http.ResponseWriter, req *http.Request, s string) Character {
	var user User
	var result Character

	reqtoken := req.Header.Get("x-token")

	fmt.Println(reqtoken)
	err := db.QueryRow("SELECT user_id FROM authentication WHERE token = ?", reqtoken).Scan(&user.ID)
	checkErr(err)

	// クエリ実行
	rows, err := db.Query("SELECT character_id FROM character_rank WHERE character_rank = ?", s)
	checkErr(err)

	defer rows.Close()

	var characterid int
	var characteridlist []int
	for rows.Next() {
		if err := rows.Scan(&characterid); err != nil {
			log.Fatal(err)
		}
		characteridlist = append(characteridlist, characterid)

	}
	if err := rows.Err(); err != nil {
		log.Fatal(err)
	}

	characterid = rand.Intn(4)

	err = db.QueryRow("SELECT character_id, character_name FROM `character` WHERE character_id = ?", characteridlist[characterid]).Scan(&result.ID, &result.Name)
	checkErr(err)

	stmt, err := db.Prepare("INSERT INTO possession_characters(user_id, character_id, character_name, issued_at) VALUES(?,?,?,?)")
	checkErr(err)
	t := time.Now()
	const layout = "2006-01-02 15:04:05"
	t.Format(layout)
	res, err := stmt.Exec(user.ID, result.ID, result.Name, t)
	checkErr(err)
	id, err := res.LastInsertId()
	checkErr(err)

	fmt.Println(characteridlist[characterid])
	fmt.Println(id)

	return result

}
