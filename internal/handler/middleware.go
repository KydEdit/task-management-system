package handler

import (
	"context"
	"net/http"
	"strings"
)

type contextKey string

const userContextKey = contextKey("userEmail")

func AuthMiddleware(
	auth AuthService,
	next http.HandlerFunc,
) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		authHeader := r.Header.Get("Authorization")
		if authHeader == "" {
			writeError(w, http.StatusUnauthorized, "missing authorization header")
			return
		}

		parts := strings.Split(authHeader, " ")
		if len(parts) != 2 || parts[0] != "Bearer" {
			writeError(w, http.StatusUnauthorized, "invalid authorization header")
			return
		}
		tokenString := parts[1]

		email, err := auth.ParseToken(tokenString)
		if err != nil {
			writeError(w, http.StatusUnauthorized, "invalid token")
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
