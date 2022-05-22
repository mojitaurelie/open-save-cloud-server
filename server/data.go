package server

import (
	"encoding/json"
	"github.com/go-chi/chi/v5"
	"io"
	"log"
	"net/http"
	"opensavecloudserver/database"
	"strconv"
	"time"
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

func GameInfoByID(w http.ResponseWriter, r *http.Request) {
	userId, err := userIdFromContext(r.Context())
	if err != nil {
		internalServerError(w, r)
		return
	}
	queryId := chi.URLParam(r, "id")
	id, err := strconv.Atoi(queryId)
	if err != nil {
		badRequest("Game ID missing or not an int", w, r)
		return
	}
	game, err := database.GameInfoById(userId, id)
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

func UploadSave(w http.ResponseWriter, r *http.Request) {
	userId, err := userIdFromContext(r.Context())
	if err != nil {
		internalServerError(w, r)
		return
	}
	gameId, err := gameIdFromContext(r.Context())
	if err != nil {
		internalServerError(w, r)
		return
	}
	defer database.UnlockGame(gameId)
	game, err := database.GameInfoById(userId, gameId)
	if err != nil {
		internalServerError(w, r)
		return
	}
	file, _, err := r.FormFile("file")
	if err != nil {
		internalServerError(w, r)
		log.Println(err)
		return
	}
	defer file.Close()
	err = database.UploadSave(file, game)
	if err != nil {
		internalServerError(w, r)
		log.Println(err)
		return
	}
	err = database.UpdateGameRevision(game)
	if err != nil {
		internalServerError(w, r)
		log.Println(err)
		return
	}
	payload := &successMessage{
		Message:   "Game uploaded",
		Timestamp: time.Now(),
		Status:    200,
	}
	ok(payload, w, r)
}
