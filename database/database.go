package database

import (
	"fmt"
	"github.com/google/uuid"
	"golang.org/x/crypto/bcrypt"
	"gorm.io/driver/mysql"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
	"log"
	"opensavecloudserver/config"
	"os"
	"time"
)

var db *gorm.DB

const AdminRole string = "admin"
const UserRole string = "user"

func init() {
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
}

func AllUsers() ([]*User, error) {
	var users []*User
	err := db.Model(User{}).Find(&users).Error
	if err != nil {
		return nil, err
	}
	for _, user := range users {
		if user.Role == AdminRole {
			user.IsAdmin = true
		}
	}
	return users, nil
}

// UserByUsername get a user by the username
func UserByUsername(username string) (*User, error) {
	var user *User
	err := db.Model(User{}).Where(User{Username: username}).First(&user).Error
	if err != nil {
		return nil, err
	}
	if user.Role == AdminRole {
		user.IsAdmin = true
	}
	return user, nil
}

// UserById get a user
func UserById(userId int) (*User, error) {
	var user *User
	err := db.Model(User{}).Where(userId).First(&user).Error
	if err != nil {
		return nil, err
	}
	if user.Role == AdminRole {
		user.IsAdmin = true
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

func SaveUser(user *User) error {
	return db.Save(user).Error
}

func RemoveUser(user *User) error {
	return db.Delete(User{}, user.ID).Error
}

func RemoveAllUserGameEntries(user *User) error {
	return db.Delete(Game{}, Game{UserId: user.ID}).Error
}

// AddAdmin register a user and set his role to admin
/*func AddAdmin(username string, password []byte) error {
	user := &User{
		Username: username,
		Password: password,
		Role:     AdminRole,
	}
	return db.Save(user).Error
}*/

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

func UpdateGameRevision(game *Game, hash string) error {
	game.Revision += 1
	if game.Hash == nil {
		game.Hash = new(string)
	}
	*game.Hash = hash
	game.Available = true
	if game.LastUpdate == nil {
		game.LastUpdate = new(time.Time)
	}
	*game.LastUpdate = time.Now()
	err := db.Save(game).Error
	if err != nil {
		return err
	}
	return nil
}

// ChangePassword change the password of the user, the param 'password' must be the clear password
func ChangePassword(userId int, password []byte) error {
	user, err := UserById(userId)
	if err != nil {
		return err
	}
	hashedPassword, err := bcrypt.GenerateFromPassword(password, *config.Features().PasswordHashCost)
	if err != nil {
		return err
	}
	user.Password = hashedPassword
	err = db.Save(user).Error
	if err != nil {
		return err
	}
	return nil
}
