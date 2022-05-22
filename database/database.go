package database

import (
	"crypto/md5"
	"encoding/hex"
	"errors"
	"fmt"
	"github.com/google/uuid"
	"gorm.io/driver/mysql"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
	"io"
	"log"
	"mime/multipart"
	"opensavecloudserver/config"
	"os"
	"path"
	"sync"
	"time"
)

var (
	locks map[int]GameUploadToken
	mu    sync.Mutex
)

var db *gorm.DB

func init() {
	locks = make(map[int]GameUploadToken)
	dbConfig := config.Database()
	var err error
	connectionString := ""
	if dbConfig.Password != nil {
		connectionString = fmt.Sprintf("%s:%s@tcp(%s:%d)/osc?charset=utf8mb4&parseTime=True&loc=Local",
			dbConfig.Username,
			*dbConfig.Password,
			dbConfig.Host,
			dbConfig.Port)
	} else {
		connectionString = fmt.Sprintf("%s@tcp(%s:%d)/osc?charset=utf8mb4&parseTime=True&loc=Local",
			dbConfig.Username,
			dbConfig.Host,
			dbConfig.Port)
	}
	db, err = gorm.Open(mysql.Open(connectionString), &gorm.Config{
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
	go func() {
		for {
			time.Sleep(time.Minute)
			clearLocks()
		}
	}()
}

// UserByUsername get a user by the username
func UserByUsername(username string) (*User, error) {
	var user *User
	err := db.Model(User{}).Where(User{Username: username}).First(&user).Error
	if err != nil {
		return nil, err
	}
	return user, nil
}

// UserById get a user
func UserById(userId int) (*User, error) {
	var user *User
	err := db.Model(User{}).Where(User{ID: userId}).First(&user).Error
	if err != nil {
		return nil, err
	}
	return user, nil
}

// AddUser register a user
func AddUser(username string, password []byte) error {
	user := &User{
		Username: username,
		Password: password,
	}
	return db.Save(user).Error
}

// GameInfoById return information of a game
func GameInfoById(userId, gameId int) (*Game, error) {
	var game *Game
	err := db.Model(Game{}).Where(Game{ID: gameId, UserId: userId}).First(&game).Error
	if err != nil {
		return nil, err
	}
	return game, nil
}

// GameInfosByUserId get all saved games for a user
func GameInfosByUserId(userId int) ([]*Game, error) {
	var games []*Game
	err := db.Model(Game{}).Where(Game{UserId: userId}).Find(&games).Error
	if err != nil {
		return nil, err
	}
	return games, nil
}

// CreateGame create an entry for a new game save, do this only for create a new entry
func CreateGame(userId int, name string) (*Game, error) {
	gameUUID := uuid.New()
	game := &Game{
		Name:        name,
		Revision:    0,
		PathStorage: gameUUID.String() + ".bin",
		UserId:      userId,
		Available:   false,
	}
	if err := db.Save(&game).Error; err != nil {
		return nil, err
	}
	return game, nil
}

// AskForUpload Create a lock for upload a new revision of a game
func AskForUpload(userId, gameId int) (*GameUploadToken, error) {
	mu.Lock()
	defer mu.Unlock()
	_, err := GameInfoById(userId, gameId)
	if err != nil {
		return nil, err
	}
	if _, ok := locks[gameId]; !ok {
		token := uuid.New()
		lock := GameUploadToken{
			GameId:      gameId,
			UploadToken: token.String(),
		}
		locks[gameId] = lock
		return &lock, nil
	}
	return nil, errors.New("game already locked")
}

func CheckUploadToken(uploadToken string) (int, bool) {
	mu.Lock()
	defer mu.Unlock()
	for _, lock := range locks {
		if lock.UploadToken == uploadToken {
			return lock.GameId, true
		}
	}
	return -1, false
}

func UploadSave(file multipart.File, game *Game) error {
	filePath := path.Join(config.Path().Storage, game.PathStorage)
	f, err := os.OpenFile(filePath, os.O_WRONLY|os.O_CREATE, 0666)
	if err != nil {
		return err
	}
	defer f.Close()
	_, err = io.Copy(f, file)
	if err != nil {
		return err
	}
	return nil
}

func UpdateGameRevision(game *Game) error {
	filePath := path.Join(config.Path().Storage, game.PathStorage)
	file, err := os.Open(filePath)
	if err != nil {
		return err
	}
	defer file.Close()

	hash := md5.New()
	_, err = io.Copy(hash, file)
	if err != nil {
		return err
	}
	sum := hash.Sum(nil)
	game.Revision += 1
	if game.Hash == nil {
		game.Hash = new(string)
	}
	*game.Hash = hex.EncodeToString(sum)
	game.Available = true
	if game.LastUpdate == nil {
		game.LastUpdate = new(time.Time)
	}
	*game.LastUpdate = time.Now()
	err = db.Save(game).Error
	if err != nil {
		return err
	}
	return nil
}

func UnlockGame(gameId int) {
	mu.Lock()
	defer mu.Unlock()
	delete(locks, gameId)
}

// clearLocks clear lock of zombi upload
func clearLocks() {
	mu.Lock()
	defer mu.Unlock()
	now := time.Now()
	toUnlock := make([]int, 0)
	for gameId, lock := range locks {
		if lock.Expire.After(now) {
			toUnlock = append(toUnlock, gameId)
		}
	}
	for _, gameId := range toUnlock {
		delete(locks, gameId)
	}
}
