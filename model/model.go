package model

import (
	"log"
	"os"
)

type UserHandler struct{}

var (
	Warn = log.New(os.Stderr, "[WARN] ", log.Ldate|log.Ltime|log.Lshortfile)
	Info = log.New(os.Stdout, "[INFO] ", log.Ldate|log.Ltime|log.Lshortfile)
)

type User struct {
	ID   string `json:"id"`
	Name string `json:"name"`
}

type Token struct {
	Token string `json:"token"`
}

type Character struct {
	ID   string `json:"characterID"`
	Name string `json:"name"`
}

type Results struct {
	Results []Character `json:"results"`
}

type Possession_character struct {
	UserID      string `json:"userCharacterID"`
	CharacterID string `json:"characterID"`
	Name        string `json:"name"`
}

type Possession_characters struct {
	Characters []Possession_character `json:"characters"`
}

type GachaTime struct {
	Time int `json:"times"`
}
