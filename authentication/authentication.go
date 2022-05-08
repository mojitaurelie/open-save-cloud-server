package authentication

import (
	"crypto/rand"
	"errors"
	"github.com/golang-jwt/jwt"
	"golang.org/x/crypto/bcrypt"
	"log"
	"opensavecloudserver/database"
	"time"
)

var secret []byte

type AccessToken struct {
	Token string `json:"token"`
}

type Registration struct {
	Username  string `json:"username"`
	Password  string `json:"password"`
	Firstname string `json:"firstname"`
	Lastname  string `json:"lastname"`
	Pronouns  int    `json:"pronouns"`
}

func init() {
	secret = make([]byte, 512)
	_, err := rand.Read(secret)
	if err != nil {
		log.Fatal(err)
	}
}

func Connect(username, password string) (*AccessToken, error) {
	user, err := database.UserByUsername(username)
	if err != nil {
		return nil, err
	}
	if err := bcrypt.CompareHashAndPassword(user.Password, []byte(password)); err != nil {
		return nil, err
	}
	token, err := token(user.ID)
	if err != nil {
		return nil, err
	}
	return &AccessToken{
		Token: token,
	}, nil
}

func ParseToken(token string) (int, error) {
	var claims jwt.MapClaims
	_, err := jwt.ParseWithClaims(token, &claims, func(token *jwt.Token) (interface{}, error) {
		return secret, nil
	})
	if err != nil {
		return 0, err
	}
	if userId, ok := claims["sub"]; ok {
		return int(userId.(float64)), nil
	}
	return 0, errors.New("this token does not have a userId in it")
}

func Register(user *Registration) error {
	_, err := database.UserByUsername(user.Username)
	if err == nil {
		return errors.New("this username already exist")
	}
	hash, err := bcrypt.GenerateFromPassword([]byte(user.Password), 12)
	if err != nil {
		return err
	}
	return database.AddUser(user.Username, user.Firstname, user.Lastname, hash, user.Pronouns)
}

func token(userId int) (string, error) {
	token := jwt.NewWithClaims(jwt.SigningMethodHS512, jwt.MapClaims{
		"sub": userId,
		"exp": time.Now().Add(1 * time.Hour).Unix(),
	})
	return token.SignedString(secret)
}
