package database

import "time"

type User struct {
	Username string `json:"username"`
	Role     string `json:"role"`
	Password []byte `json:"-"`
	ID       int    `json:"id"`
	IsAdmin  bool   `json:"is_admin" gorm:"-:all"`
}

type Game struct {
	Name        string     `json:"name"`
	PathStorage string     `json:"-"`
	ID          int        `json:"id"`
	Revision    int        `json:"rev"`
	UserId      int        `json:"-"`
	Available   bool       `json:"available"`
	Hash        *string    `json:"hash"`
	LastUpdate  *time.Time `json:"last_update"`
}
