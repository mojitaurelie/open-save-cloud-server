package upload

import (
	"crypto/sha512"
	"errors"
	"fmt"
	"github.com/google/uuid"
	"io"
	"log"
	"mime/multipart"
	"opensavecloudserver/config"
	"opensavecloudserver/database"
	"os"
	"path"
	"strconv"
	"sync"
	"time"
)

var (
	locks map[int]GameUploadToken
	mu    sync.Mutex
)

type GameUploadToken struct {
	GameId      int       `json:"-"`
	UploadToken string    `json:"upload_token"`
	Expire      time.Time `json:"expire"`
}

func init() {
	locks = make(map[int]GameUploadToken)
	go func() {
		for {
			time.Sleep(time.Minute)
			clearLocks()
		}
	}()
}

func MoveFile(sourcePath, destPath string) error {
	inputFile, err := os.Open(sourcePath)
	if err != nil {
		return fmt.Errorf("couldn't open source file: %s", err)
	}
	outputFile, err := os.Create(destPath)
	if err != nil {
		err := inputFile.Close()
		if err != nil {
			return err
		}
		return fmt.Errorf("couldn't open dest file: %s", err)
	}
	defer func(outputFile *os.File) {
		err := outputFile.Close()
		if err != nil {
			log.Println(err)
		}
	}(outputFile)
	_, err = io.Copy(outputFile, inputFile)
	if err != nil {
		err := inputFile.Close()
		if err != nil {
			return err
		}
		return fmt.Errorf("writing to output file failed: %s", err)
	}
	err = inputFile.Close()
	if err != nil {
		return err
	}
	// The copy was successful, so now delete the original file
	err = os.Remove(sourcePath)
	if err != nil {
		return fmt.Errorf("failed removing original file: %s", err)
	}
	return nil
}

// AskForUpload Create a lock for upload a new revision of a game
func AskForUpload(userId, gameId int) (*GameUploadToken, error) {
	mu.Lock()
	defer mu.Unlock()
	_, err := database.GameInfoById(userId, gameId)
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

func UploadToCache(file multipart.File, game *database.Game) error {
	filePath := path.Join(config.Path().Cache, strconv.Itoa(game.UserId))
	if _, err := os.Stat(filePath); err != nil {
		err = os.Mkdir(filePath, 0766)
		if err != nil {
			return err
		}
	}
	filePath = path.Join(filePath, game.PathStorage)
	f, err := os.OpenFile(filePath, os.O_WRONLY|os.O_CREATE, 0666)
	if err != nil {
		return err
	}
	defer func(f *os.File) {
		err := f.Close()
		if err != nil {
			log.Println(err)
		}
	}(f)
	if _, err := io.Copy(f, file); err != nil {
		return err
	}
	return nil
}

func ValidateAndMove(game *database.Game, hash string) error {
	filePath := path.Join(config.Path().Cache, strconv.Itoa(game.UserId), game.PathStorage)
	if err := checkHash(filePath, hash); err != nil {
		return err
	}
	if err := moveToStorage(filePath, game); err != nil {
		return err
	}
	return nil
}

func checkHash(path, hash string) error {
	h, err := FileHash(path)
	if err != nil {
		return err
	}
	if h != hash {
		return errors.New("hash is different")
	}
	return nil
}

func moveToStorage(cachePath string, game *database.Game) error {
	filePath := path.Join(config.Path().Storage, strconv.Itoa(game.UserId))
	if _, err := os.Stat(filePath); err != nil {
		err = os.Mkdir(filePath, 0766)
		if err != nil {
			return err
		}
	}
	filePath = path.Join(filePath, game.PathStorage)
	if _, err := os.Stat(filePath); err == nil {
		if err = os.Remove(filePath); err != nil {
			return err
		}
	}
	if err := MoveFile(cachePath, filePath); err != nil {
		return err
	}
	return nil
}

func UnlockGame(gameId int) {
	mu.Lock()
	defer mu.Unlock()
	delete(locks, gameId)
}

// RemoveFolders remove all files of the user from storage and cache
func RemoveFolders(userId int) error {
	userPath := path.Join(config.Path().Storage, strconv.Itoa(userId))
	userCache := path.Join(config.Path().Cache, strconv.Itoa(userId))
	if _, err := os.Stat(userPath); err == nil {
		err := os.RemoveAll(userPath)
		if err != nil {
			log.Fatal(err)
		}
	}
	if _, err := os.Stat(userCache); err == nil {
		err := os.RemoveAll(userCache)
		if err != nil {
			log.Fatal(err)
		}
	}
	return nil
}

func RemoveGame(userId int, game *database.Game) error {
	filePath := path.Join(config.Path().Storage, strconv.Itoa(userId), game.PathStorage)
	return os.Remove(filePath)
}

func FileHash(path string) (string, error) {
	f, err := os.Open(path)
	if err != nil {
		return "", err
	}
	defer func(f *os.File) {
		err := f.Close()
		if err != nil {
			log.Println(err)
		}
	}(f)
	h := sha512.New()
	if _, err := io.Copy(h, f); err != nil {
		return "", err
	}
	return fmt.Sprintf("%x", h.Sum(nil)), nil
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
