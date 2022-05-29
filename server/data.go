package server

import (
	"encoding/json"
	"github.com/go-chi/chi/v5"
	"io"
	"log"
	"net/http"
	"opensavecloudserver/config"
	"opensavecloudserver/database"
	"opensavecloudserver/upload"
	"os"
	"path/filepath"
	"strconv"
	"time"
	"unicode/utf8"
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

type NewPassword struct {
	Password       string `json:"password"`
	VerifyPassword string `json:"verify_password"`
}

// CreateGame create a game entry to the database
func CreateGame(w http.ResponseWriter, r *http.Request) {
	userId, err := userIdFromContext(r.Context())
	if err != nil {
		internalServerError(w, r)
		log.Println(err)
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

// GameInfoByID get the game save information from the database
func GameInfoByID(w http.ResponseWriter, r *http.Request) {
	userId, err := userIdFromContext(r.Context())
	if err != nil {
		internalServerError(w, r)
		log.Println(err)
		return
	}
	queryId := chi.URLParam(r, "id")
	id, err := strconv.Atoi(queryId)
	if err != nil {
		badRequest("Game ID missing or not an int", w, r)
		log.Println(err)
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

// AllGamesInformation all game saves information for a user
func AllGamesInformation(w http.ResponseWriter, r *http.Request) {
	userId, err := userIdFromContext(r.Context())
	if err != nil {
		internalServerError(w, r)
		log.Println(err)
		return
	}
	games, err := database.GameInfosByUserId(userId)
	if err != nil {
		internalServerError(w, r)
		log.Println(err)
		return
	}
	ok(games, w, r)
}

// AskForUpload check if the game save is not lock, then lock it and generate a token
func AskForUpload(w http.ResponseWriter, r *http.Request) {
	userId, err := userIdFromContext(r.Context())
	if err != nil {
		internalServerError(w, r)
		log.Println(err)
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
	token, err := upload.AskForUpload(userId, gameInfo.GameId)
	if err != nil {
		ok(LockError{Message: err.Error()}, w, r)
		return
	}
	ok(token, w, r)
}

// UploadSave upload the game save archive to the storage folder
func UploadSave(w http.ResponseWriter, r *http.Request) {
	userId, err := userIdFromContext(r.Context())
	if err != nil {
		internalServerError(w, r)
		log.Println(err)
		return
	}
	gameId, err := gameIdFromContext(r.Context())
	if err != nil {
		internalServerError(w, r)
		log.Println(err)
		return
	}
	defer upload.UnlockGame(gameId)
	hash := r.Header.Get("X-Game-Save-Hash")
	if utf8.RuneCountInString(hash) == 0 {
		badRequest("The header X-Game-Save-Hash is missing", w, r)
		return
	}
	game, err := database.GameInfoById(userId, gameId)
	if err != nil {
		internalServerError(w, r)
		log.Println(err)
		return
	}
	file, _, err := r.FormFile("file")
	if err != nil {
		internalServerError(w, r)
		log.Println(err)
		return
	}
	defer file.Close()
	err = upload.UploadSave(file, game)
	if err != nil {
		internalServerError(w, r)
		log.Println(err)
		return
	}
	err = database.UpdateGameRevision(game, hash)
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

// Download send the game save archive to the client
func Download(w http.ResponseWriter, r *http.Request) {
	userId, err := userIdFromContext(r.Context())
	if err != nil {
		internalServerError(w, r)
		log.Println(err)
		return
	}
	gameId, err := gameIdFromContext(r.Context())
	if err != nil {
		internalServerError(w, r)
		log.Println(err)
		return
	}
	defer upload.UnlockGame(gameId)
	game, err := database.GameInfoById(userId, gameId)
	if err != nil {
		internalServerError(w, r)
		log.Println(err)
		return
	}
	savePath := filepath.Join(config.Path().Storage, game.PathStorage)

	if _, err := os.Stat(savePath); err == nil {
		file, err := os.Open(savePath)
		if err != nil {
			internalServerError(w, r)
			log.Println(err)
			return
		}
		defer file.Close()
		_, err = io.Copy(w, file)
		if err != nil {
			internalServerError(w, r)
			log.Println(err)
			return
		}
	} else {
		http.NotFound(w, r)
	}
}

func UserInformation(w http.ResponseWriter, r *http.Request) {
	userId, err := userIdFromContext(r.Context())
	if err != nil {
		internalServerError(w, r)
		log.Println(err)
		return
	}
	user, err := database.UserById(userId)
	if err != nil {
		internalServerError(w, r)
		log.Println(err)
		return
	}
	ok(user, w, r)
}

func ChangePassword(w http.ResponseWriter, r *http.Request) {
	userId, err := userIdFromContext(r.Context())
	if err != nil {
		internalServerError(w, r)
		log.Println(err)
		return
	}
	body, err := io.ReadAll(r.Body)
	if err != nil {
		internalServerError(w, r)
		log.Println(err)
		return
	}
	newPassword := new(NewPassword)
	err = json.Unmarshal(body, newPassword)
	if err != nil {
		internalServerError(w, r)
		log.Println(err)
		return
	}
	if newPassword.Password != newPassword.VerifyPassword {
		badRequest("password are not the same", w, r)
		return
	}
	err = database.ChangePassword(userId, []byte(newPassword.Password))
	if err != nil {
		internalServerError(w, r)
		log.Println(err)
		return
	}
	payload := &successMessage{
		Message:   "Password changed",
		Timestamp: time.Now(),
		Status:    200,
	}
	ok(payload, w, r)
}
