package upload

import (
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

func ProcessFile(file multipart.File, game *database.Game) error {
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
	_, err = io.Copy(f, file)
	if err != nil {
		return err
	}
	err = moveToStorage(filePath, game)
	if err != nil {
		return err
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
