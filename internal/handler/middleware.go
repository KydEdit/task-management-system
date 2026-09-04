package handler

import (
	"context"
	"net/http"
	"strings"
	"task-manager-api/internal/service"
)

type contextKey string

const userContextKey = contextKey("userEmail")

func AuthMiddleware(
	auth *service.AuthService,
	next http.HandlerFunc,
) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		authHeader := r.Header.Get("Authorization")
		if authHeader == "" {
			http.Error(
				w,
				"Missing Authorization header",
				http.StatusUnauthorized,
			)
			return
		}

		parts := strings.Split(authHeader, " ")
		if len(parts) != 2 {
			http.Error(
				w,
				"Invalid Authorization format",
				http.StatusUnauthorized,
			)
			return
		}

		tokenString := parts[1]

		email, err := auth.ParseToken(tokenString)
		if err != nil {
			http.Error(
				w,
				"Invalid token",
				http.StatusUnauthorized,
			)
			return
		}

		ctx := context.WithValue(
			r.Context(),
			userContextKey,
			email,
		)

		next(
			w,
			r.WithContext(ctx),
		)
	}
}
