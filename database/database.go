package database

import (
	"fmt"
	"gorm.io/driver/mysql"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
	"log"
	"opensavecloudserver/config"
	"os"
	"time"
)

var db *gorm.DB

func init() {
	dbConfig := config.Database()
	var err error
	db, err = gorm.Open(mysql.Open(
		fmt.Sprintf("%s:%s@tcp(%s:%d)/transagenda?charset=utf8mb4&parseTime=True&loc=Local",
			dbConfig.Username,
			dbConfig.Password,
			dbConfig.Host,
			dbConfig.Port),
	), &gorm.Config{
		Logger: logger.New(
			log.New(os.Stdout, "", log.LstdFlags), // io writer
			logger.Config{
				SlowThreshold:             time.Second,  // Slow SQL threshold
				LogLevel:                  logger.Error, // Log level
				IgnoreRecordNotFoundError: true,         // Ignore ErrRecordNotFound error for logger
				Colorful:                  true,         // Enable color
			},
		),
	})
	if err != nil {
		log.Fatal(err)
	}
}

func UserByUsername(username string) (*User, error) {
	var user *User
	err := db.Model(User{}).Where(User{Username: username}).First(&user).Error
	if err != nil {
		return nil, err
	}
	return user, nil
}

func UserById(userId int) (*User, error) {
	var user *User
	err := db.Model(User{}).Where(User{ID: userId}).First(&user).Error
	if err != nil {
		return nil, err
	}
	return user, nil
}

func AddUser(username string, password []byte) error {
	user := &User{
		Username: username,
		Password: password,
	}
	return db.Save(user).Error
}

func GameInfoById(userId, gameId int) (*Game, error) {
	var game *Game
	err := db.Model(Game{}).Where(Game{ID: gameId, UserId: userId}).First(&game).Error
	if err != nil {
		return nil, err
	}
	return game, nil
}

func GameInfosByUserId(userId int) ([]*Game, error) {
	var games []*Game
	err := db.Model(Game{}).Where(Game{UserId: userId}).Find(&games).Error
	if err != nil {
		return nil, err
	}
	return games, nil
}
