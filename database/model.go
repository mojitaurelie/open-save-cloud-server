package database

import "time"

type User struct {
	ID       int    `json:"id"`
	Username string `json:"username"`
	Role     string `json:"role"`
	Password []byte `json:"-"`
	IsAdmin  bool   `json:"is_admin" gorm:"-:all"`
}

type Game struct {
	ID          int        `json:"id"`
	Name        string     `json:"name"`
	Revision    int        `json:"rev"`
	PathStorage string     `json:"-"`
	Hash        *string    `json:"hash"`
	LastUpdate  *time.Time `json:"last_update"`
	UserId      int        `json:"-"`
	Available   bool       `json:"available"`
}
