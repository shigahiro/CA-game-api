package handler

import (
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"math/rand"
	"net/http"
	"time"

	"github.com/shigahiro/CA-game-api/model"
)

func Gachadraw(w http.ResponseWriter, req *http.Request) {
	model.Info.Println("ガチャ実行ルーティング成功")

	var results model.Results
	var gacha_times model.GachaTime

	var i interface{} = &gacha_times
	if err := unmarshalingjson(i, w, req); err != nil {
		return
	}

	db := db_open()
	defer db.Close()
	for count := gacha_times.Time; count >= 1; count-- {
		randam := rand.Intn(100)
		switch {
		case 0 <= randam && randam < 4:
			s := "S"
			result, err := gacha_data_insert(db, w, req, s)
			if err != nil {
				return
			}
			results.Results = append(results.Results, result)
		case 4 <= randam && randam < 20:
			s := "A"
			result, err := gacha_data_insert(db, w, req, s)
			if err != nil {
				return
			}
			results.Results = append(results.Results, result)
		default:
			s := "B"
			result, err := gacha_data_insert(db, w, req, s)
			if err != nil {
				return
			}
			results.Results = append(results.Results, result)
		}
	}

	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(results)
}

func gacha_data_insert(db *sql.DB, w http.ResponseWriter, req *http.Request, s string) (model.Character, error) {
	var user model.User
	var result model.Character

	reqtoken := req.Header.Get("x-token")
	if reqtoken == "" {
		model.Warn.Println("トークンを取得できません")
		return result, errors.New("トークンを取得できません")
	}
	model.Info.Println("トークン取得成功")

	err := db.QueryRow("SELECT id FROM users WHERE token = ?", reqtoken).Scan(&user.ID)
	if err != nil {
		model.Warn.Println("トークンがありません")
		return result, errors.New("トークンがありません")
	}
	model.Info.Println("トークンが一致しました")

	// クエリ実行
	rows, err := db.Query("SELECT id FROM characters WHERE `rank` = ?", s)
	if err != nil {
		model.Warn.Println("キャラクターID取得失敗")
		return result, errors.New("キャラクターID取得失敗")
	}
	model.Info.Println("キャラクターID取得成功")

	defer rows.Close()

	var characterid int
	var characteridlist []int
	for rows.Next() {
		if err := rows.Scan(&characterid); err != nil {
			if err != nil {
				model.Warn.Println("変数へのカラム格納失敗")
				return result, errors.New("変数へのカラム格納失敗")
			}
		}
		characteridlist = append(characteridlist, characterid)

	}
	if err := rows.Err(); err != nil {
		if err != nil {
			model.Warn.Println("正常に行セットのループ処理が終了しませんでした")
			return result, errors.New("正常に行セットのループ処理が終了しませんでした")
		}
	}

	characterid = rand.Intn(len(characteridlist))
	fmt.Println(characteridlist, characterid)

	err = db.QueryRow("SELECT id, name FROM `characters` WHERE id = ?", characteridlist[characterid]).Scan(&result.ID, &result.Name)

	if err != nil {
		model.Warn.Println("キャラクター名取得失敗")
		return result, errors.New("キャラクター名取得失敗")
	}

	stmt, err := db.Prepare("INSERT INTO UserCharacter(user_id, character_id) VALUES(?,?)")
	if err != nil {
		model.Warn.Println("Stmtオブジェクト生成失敗")
		return result, errors.New("Stmtオブジェクト生成失敗")
	}
	model.Info.Println("Stmtオブジェクト生成成功")

	t := time.Now()
	const layout = "2006-01-02 15:04:05"
	t.Format(layout)

	_, err = stmt.Exec(user.ID, result.ID)
	if err != nil {
		model.Warn.Println("キャラクター保存失敗")
		return result, errors.New("キャラクター保存失敗")
	}
	model.Info.Println("キャラクター保存成功")

	return result, nil

}

func Character_list(w http.ResponseWriter, req *http.Request) {
	model.Info.Println("ユーザ所持キャラ一覧取得ルーティング成功")

	var id int
	reqtoken := req.Header.Get("x-token")
	if reqtoken == "" {
		model.Warn.Println("トークンを取得できません")
		return
	}

	db := db_open()
	defer db.Close()
	err := db.QueryRow("SELECT id FROM users WHERE token = ?", reqtoken).Scan(&id)
	if err != nil {
		model.Warn.Println("トークンがありません")
		return
	}
	model.Info.Println("トークンが一致しました")

	rows, err := db.Query("select UserCharacter.user_id, UserCharacter.character_id, characters.name from UserCharacter inner join characters on UserCharacter.character_id = characters.id where UserCharacter.user_id = ?", id)

	if err != nil {
		model.Warn.Println("所持キャラクター取得失敗")
		return
	}
	model.Info.Println("所持キャラクター取得成功")

	defer rows.Close()

	var possession_character model.Possession_character
	var possession_characters model.Possession_characters
	for rows.Next() {
		if err := rows.Scan(&possession_character.UserID, &possession_character.CharacterID, &possession_character.Name); err != nil {
			if err != nil {
				model.Warn.Println("変数へのカラム格納失敗")
				return
			}
		}
		possession_characters.Characters = append(possession_characters.Characters, possession_character)

	}
	if err := rows.Err(); err != nil {
		if err != nil {
			model.Warn.Println("正常に行セットのループ処理が終了しませんでした")
			return
		}
	}

	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(possession_characters)
}
