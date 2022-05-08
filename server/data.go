package server

import (
	"encoding/json"
	"io"
	"log"
	"net/http"
	"opensavecloudserver/database"
)

type NewGameInfo struct {
	Name string `json:"name"`
}

type UploadGameInfo struct {
	GameId int `json:"game_id"`
}

type LockError struct {
	Message string `json:"message"`
}

func CreateGame(w http.ResponseWriter, r *http.Request) {
	userId, err := userIdFromContext(r.Context())
	if err != nil {
		internalServerError(w, r)
		return
	}
	body, err := io.ReadAll(r.Body)
	if err != nil {
		internalServerError(w, r)
		log.Println(err)
		return
	}
	gameInfo := new(NewGameInfo)
	err = json.Unmarshal(body, gameInfo)
	if err != nil {
		internalServerError(w, r)
		log.Println(err)
		return
	}
	game, err := database.CreateGame(userId, gameInfo.Name)
	if err != nil {
		internalServerError(w, r)
		log.Println(err)
		return
	}
	ok(game, w, r)
}

func AskForUpload(w http.ResponseWriter, r *http.Request) {
	userId, err := userIdFromContext(r.Context())
	if err != nil {
		internalServerError(w, r)
		return
	}
	body, err := io.ReadAll(r.Body)
	if err != nil {
		internalServerError(w, r)
		log.Println(err)
		return
	}
	gameInfo := new(UploadGameInfo)
	err = json.Unmarshal(body, gameInfo)
	if err != nil {
		internalServerError(w, r)
		log.Println(err)
		return
	}
	token, err := database.AskForUpload(userId, gameInfo.GameId)
	if err != nil {
		ok(LockError{Message: err.Error()}, w, r)
		return
	}
	ok(token, w, r)
}
