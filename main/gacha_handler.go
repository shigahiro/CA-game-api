package main

import (
	"database/sql"
	"encoding/json"
	"math/rand"
	"net/http"
	"time"
)

func gachadraw(db *sql.DB, w http.ResponseWriter, req *http.Request) {
	var results Results
	var gacha_times GachaTime
	var judge Character

	var i interface{}
	i = &gacha_times
	if err := unmarshalingjson(i, w, req); err != nil {
		return
	}

	for count := gacha_times.Time; count >= 1; count-- {
		randam := rand.Intn(100)
		switch {
		case 0 <= randam && randam < 4:
			s := "S"
			result := gacha_data_insert(db, w, req, s)
			if result == judge {
				return
			}
			results.Results = append(results.Results, result)
		case 4 <= randam && randam < 20:
			s := "A"
			result := gacha_data_insert(db, w, req, s)
			if result == judge {
				return
			}
			results.Results = append(results.Results, result)
		default:
			s := "B"
			result := gacha_data_insert(db, w, req, s)
			if result == judge {
				return
			}
			results.Results = append(results.Results, result)
		}
	}

	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(results)
}

func gacha_data_insert(db *sql.DB, w http.ResponseWriter, req *http.Request, s string) Character {
	var user User
	var result Character

	reqtoken := req.Header.Get("x-token")
	if reqtoken == "" {
		warn.Println("トークンを取得できません")
		return result
	}
	info.Println("トークン取得成功")

	err := db.QueryRow("SELECT user_id FROM authentication WHERE token = ?", reqtoken).Scan(&user.ID)
	checkErr(err, "トークンがありません")
	info.Println("トークンが一致しました")

	// クエリ実行
	rows, err := db.Query("SELECT character_id FROM character_rank WHERE character_rank = ?", s)
	checkErr(err, "キャラクターIDID取得失敗")
	info.Println("キャラクターID取得成功")

	defer rows.Close()

	var characterid int
	var characteridlist []int
	for rows.Next() {
		if err := rows.Scan(&characterid); err != nil {
			checkErr(err, "変数へのカラム格納失敗")
		}
		characteridlist = append(characteridlist, characterid)

	}
	if err := rows.Err(); err != nil {
		checkErr(err, "正常に行セットのループ処理が終了しませんでした")
	}

	characterid = rand.Intn(4)

	err = db.QueryRow("SELECT character_id, character_name FROM `character` WHERE character_id = ?", characteridlist[characterid]).Scan(&result.ID, &result.Name)
	checkErr(err, "キャラクター名取得失敗")
	info.Println("キャラクター名取得成功")

	stmt, err := db.Prepare("INSERT INTO possession_characters(user_id, character_id, character_name, issued_at) VALUES(?,?,?,?)")
	checkErr(err, "Stmtオブジェクト生成失敗")
	info.Println("Stmtオブジェクト生成成功")
	t := time.Now()
	const layout = "2006-01-02 15:04:05"
	t.Format(layout)
	_, err = stmt.Exec(user.ID, result.ID, result.Name, t)
	checkErr(err, "キャラクター保存失敗")
	info.Println("キャラクター保存失敗")

	return result

}

func character_list_get(db *sql.DB, w http.ResponseWriter, req *http.Request) {
	var id int
	reqtoken := req.Header.Get("x-token")
	if reqtoken == "" {
		warn.Println("トークンを取得できません")
		return
	}

	err := db.QueryRow("SELECT user_id FROM authentication WHERE token = ?", reqtoken).Scan(&id)
	checkErr(err, "トークンがありません")
	info.Println("トークンが一致しました")

	rows, err := db.Query("SELECT user_id, character_id, character_name FROM possession_characters WHERE user_id = ?", id)
	checkErr(err, "所持キャラクター取得失敗")
	info.Println("所持キャラクター取得成功")

	defer rows.Close()

	var possession_character Possession_character
	var possession_characters Possession_characters
	for rows.Next() {
		if err := rows.Scan(&possession_character.UserID, &possession_character.CharacterID, &possession_character.Name); err != nil {
			checkErr(err, "変数へのカラム格納失敗")
		}
		possession_characters.Characters = append(possession_characters.Characters, possession_character)

	}
	if err := rows.Err(); err != nil {
		checkErr(err, "正常に行セットのループ処理が終了しませんでした")
	}

	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(possession_characters)
}
