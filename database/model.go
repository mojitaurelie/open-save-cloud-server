package database

import "time"

type User struct {
	ID       int    `json:"id"`
	Username string `json:"username"`
	Password []byte `json:"-"`
}

type Game struct {
	ID          int       `json:"id"`
	Name        string    `json:"name"`
	Revision    int       `json:"rev"`
	PathStorage string    `json:"-"`
	Hash        string    `json:"hash"`
	LastUpdate  time.Time `json:"last_update"`
	UserId      int       `json:"-"`
	Available   bool      `json:"available"`
}
