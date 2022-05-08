package server

import (
	"context"
	"errors"
	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"log"
	"net/http"
	"opensavecloudserver/authentication"
	"opensavecloudserver/config"
)

type ContextKey string

const UserIdKey ContextKey = "userId"

// Serve start the http server
func Serve() {
	router := chi.NewRouter()
	router.Use(middleware.Logger)
	router.Use(recovery)
	router.Route("/api", func(rApi chi.Router) {
		rApi.Route("/v1", func(r chi.Router) {
			r.Post("/login", Login)
			if config.Features().AllowRegister {
				r.Post("/register", Register)
			}
			r.Route("/system", func(systemRouter chi.Router) {
				systemRouter.Get("/information", Information)
			})
			r.Group(func(secureRouter chi.Router) {
				secureRouter.Use(authMiddleware)
				secureRouter.Post("/game/create", CreateGame)
				secureRouter.Post("/game/upload/init", AskForUpload)
			})
		})
	})
	log.Println("Server is listening...")
	err := http.ListenAndServe(":8080", router)
	if err != nil {
		log.Fatal(err)
	}
}

// authMiddleware filter the request
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

func userIdFromContext(ctx context.Context) (int, error) {
	if userId, ok := ctx.Value(UserIdKey).(int); ok {
		return userId, nil
	}
	return 0, errors.New("userId not found in context")
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
