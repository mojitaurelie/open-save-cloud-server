package server

import (
	"encoding/json"
	"github.com/go-chi/chi/v5"
	"io"
	"log"
	"net/http"
	"opensavecloudserver/admin"
	"opensavecloudserver/authentication"
	"opensavecloudserver/database"
	"strconv"
	"time"
)

func AddUser(w http.ResponseWriter, r *http.Request) {
	body, err := io.ReadAll(r.Body)
	if err != nil {
		internalServerError(w, r)
		log.Println(err)
		return
	}
	userInfo := new(authentication.Registration)
	err = json.Unmarshal(body, userInfo)
	if err != nil {
		internalServerError(w, r)
		log.Println(err)
		return
	}
	err = authentication.Register(userInfo)
	if err != nil {
		internalServerError(w, r)
		log.Println(err)
		return
	}
	user, err := database.UserByUsername(userInfo.Username)
	if err != nil {
		internalServerError(w, r)
		log.Println(err)
		return
	}
	ok(user, w, r)
}

func RemoveUser(w http.ResponseWriter, r *http.Request) {
	queryId := chi.URLParam(r, "id")
	id, err := strconv.Atoi(queryId)
	if err != nil {
		internalServerError(w, r)
		log.Println(err)
		return
	}
	user, err := database.UserById(id)
	if err != nil {
		notFound(err.Error(), w, r)
		log.Println(err)
		return
	}
	err = admin.RemoveUser(user)
	if err != nil {
		internalServerError(w, r)
		log.Println(err)
		return
	}
	ok(user, w, r)
}

func AllUsers(w http.ResponseWriter, r *http.Request) {
	users, err := database.AllUsers()
	if err != nil {
		internalServerError(w, r)
		log.Println(err)
		return
	}
	ok(users, w, r)
}

func User(w http.ResponseWriter, r *http.Request) {
	queryId := chi.URLParam(r, "id")
	id, err := strconv.Atoi(queryId)
	if err != nil {
		internalServerError(w, r)
		log.Println(err)
		return
	}
	user, err := database.UserById(id)
	if err != nil {
		notFound(err.Error(), w, r)
		log.Println(err)
		return
	}
	ok(user, w, r)
}

func SetAdmin(w http.ResponseWriter, r *http.Request) {
	queryId := chi.URLParam(r, "id")
	id, err := strconv.Atoi(queryId)
	if err != nil {
		internalServerError(w, r)
		log.Println(err)
		return
	}
	user, err := database.UserById(id)
	if err != nil {
		notFound(err.Error(), w, r)
		log.Println(err)
		return
	}
	err = admin.SetAdmin(user)
	if err != nil {
		notFound(err.Error(), w, r)
		log.Println(err)
		return
	}
	ok(user, w, r)
}

func SetNotAdmin(w http.ResponseWriter, r *http.Request) {
	queryId := chi.URLParam(r, "id")
	id, err := strconv.Atoi(queryId)
	if err != nil {
		internalServerError(w, r)
		log.Println(err)
		return
	}
	user, err := database.UserById(id)
	if err != nil {
		notFound(err.Error(), w, r)
		log.Println(err)
		return
	}
	err = admin.RemoveAdminRole(user)
	if err != nil {
		notFound(err.Error(), w, r)
		log.Println(err)
		return
	}
	ok(user, w, r)
}

func ChangeUserPassword(w http.ResponseWriter, r *http.Request) {
	queryId := chi.URLParam(r, "id")
	userId, err := strconv.Atoi(queryId)
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
