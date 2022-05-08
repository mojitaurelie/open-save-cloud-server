package server

import (
	"encoding/json"
	"io"
	"log"
	"net/http"
	"opensavecloudserver/authentication"
	"time"
)

type Credential struct {
	Username string `json:"username"`
	Password string `json:"password"`
}

func Login(w http.ResponseWriter, r *http.Request) {
	body, err := io.ReadAll(r.Body)
	if err != nil {
		internalServerError(w, r)
		log.Println(err)
		return
	}
	credential := new(Credential)
	err = json.Unmarshal(body, credential)
	if err != nil {
		internalServerError(w, r)
		log.Println(err)
		return
	}
	token, err := authentication.Connect(credential.Username, credential.Password)
	if err != nil {
		unauthorized(w, r)
		return
	}
	ok(token, w, r)
}

func Register(w http.ResponseWriter, r *http.Request) {
	body, err := io.ReadAll(r.Body)
	if err != nil {
		internalServerError(w, r)
		log.Println(err)
		return
	}
	registration := new(authentication.Registration)
	err = json.Unmarshal(body, registration)
	if err != nil {
		internalServerError(w, r)
		log.Println(err)
		return
	}
	err = authentication.Register(registration)
	if err != nil {
		badRequest(err.Error(), w, r)
		return
	}
	payload := successMessage{
		Message:   "You are now registered",
		Timestamp: time.Now(),
		Status:    200,
	}
	ok(payload, w, r)
}
