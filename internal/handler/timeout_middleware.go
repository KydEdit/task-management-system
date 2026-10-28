package handler

import (
	"context"
	"net/http"
	"time"
)

func TimeoutMiddleware(d time.Duration, next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		ctx, cancel := context.WithTimeout(r.Context(), d)
		defer cancel()

		next(w, r.WithContext(ctx))
	}
}
