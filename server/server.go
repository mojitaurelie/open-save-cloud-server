package server

import (
	"context"
	"errors"
	"fmt"
	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"log"
	"net/http"
	"opensavecloudserver/authentication"
	"opensavecloudserver/config"
	"opensavecloudserver/database"
	"opensavecloudserver/upload"
)

type ContextKey string

const (
	UserIdKey ContextKey = "userId"
	GameIdKey ContextKey = "gameId"
)

// Serve start the http server
func Serve() {
	router := chi.NewRouter()
	router.Use(middleware.Logger)
	router.Use(recovery)
	router.Route("/api", func(rApi chi.Router) {
		rApi.Route("/v1", func(r chi.Router) {
			r.Post("/login", Login)
			r.Post("/check/token", CheckToken)
			if config.Features().AllowRegister {
				r.Post("/register", Register)
			}
			r.Route("/system", func(systemRouter chi.Router) {
				systemRouter.Get("/information", Information)
				systemRouter.Group(func(secureRouter chi.Router) {
					secureRouter.Get("/users", AllUsers)
				})
			})
			r.Route("/user", func(secureRouter chi.Router) {
				secureRouter.Use(authMiddleware)
				secureRouter.Get("/information", UserInformation)
				secureRouter.Post("/passwd", ChangePassword)
			})
			r.Route("/admin", func(secureRouter chi.Router) {
				secureRouter.Use(authMiddleware)
				secureRouter.Use(adminMiddleware)
			})
			r.Route("/game", func(secureRouter chi.Router) {
				secureRouter.Use(authMiddleware)
				secureRouter.Post("/create", CreateGame)
				secureRouter.Get("/all", AllGamesInformation)
				secureRouter.Get("/info/{id}", GameInfoByID)
				secureRouter.Post("/upload/init", AskForUpload)
				secureRouter.Group(func(uploadRouter chi.Router) {
					uploadRouter.Use(uploadMiddleware)
					uploadRouter.Post("/upload", UploadSave)
					uploadRouter.Get("/download", Download)
				})
			})

		})
	})
	log.Println("Server is listening...")
	err := http.ListenAndServe(fmt.Sprintf(":%d", config.Server().Port), router)
	if err != nil {
		log.Fatal(err)
	}
}

// authMiddleware check the authentication token before accessing to the resource
func authMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		header := r.Header.Get("Authorization")
		if len(header) > 7 {
			userId, err := authentication.ParseToken(header[7:])
			if err != nil {
				unauthorized(w, r)
				return
			}
			ctx := context.WithValue(r.Context(), UserIdKey, userId)
			r = r.WithContext(ctx)
			next.ServeHTTP(w, r)
			return
		}
		unauthorized(w, r)
	})
}

// adminMiddleware check the role of the user before accessing to the resource
func adminMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		header := r.Header.Get("Authorization")
		if len(header) > 7 {
			userId, err := authentication.ParseToken(header[7:])
			if err != nil {
				unauthorized(w, r)
				return
			}
			user, err := database.UserById(userId)
			if err != nil {
				internalServerError(w, r)
				log.Println(err)
				return
			}
			if !user.IsAdmin {
				forbidden(w, r)
				return
			}
			ctx := context.WithValue(r.Context(), UserIdKey, userId)
			r = r.WithContext(ctx)
			next.ServeHTTP(w, r)
			return
		}
		unauthorized(w, r)
	})
}

// uploadMiddleware check the upload key before allowing to upload a file
func uploadMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		header := r.Header.Get("X-Upload-Key")
		if len(header) > 0 {
			if gameId, ok := upload.CheckUploadToken(header); ok {
				ctx := context.WithValue(r.Context(), GameIdKey, gameId)
				r = r.WithContext(ctx)
				next.ServeHTTP(w, r)
				return
			}
		}
		unauthorized(w, r)
	})
}

func userIdFromContext(ctx context.Context) (int, error) {
	if userId, ok := ctx.Value(UserIdKey).(int); ok {
		return userId, nil
	}
	return 0, errors.New("userId not found in context")
}

func gameIdFromContext(ctx context.Context) (int, error) {
	if gameId, ok := ctx.Value(GameIdKey).(int); ok {
		return gameId, nil
	}
	return 0, errors.New("gameId not found in context")
}

func recovery(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		defer func() {
			err := recover()
			if err != nil {
				internalServerError(w, r)
			}
		}()
		next.ServeHTTP(w, r)
	})
}
