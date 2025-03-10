package middleware

import (
	"context"
	"net/http"
	"strings"
	"toptal/internal/app/auth"
	"toptal/internal/app/handler/model"
)

func JWTMiddleware(next func(w http.ResponseWriter, r *http.Request)) func(w http.ResponseWriter, r *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		authHeader := r.Header.Get("Authorization")
		if authHeader == "" {
			model.Unauthorized(w, "missing authorization header", r.URL.Path)
			return
		}

		tokenString := strings.TrimPrefix(authHeader, "Bearer ")
		if tokenString == "" {
			model.Unauthorized(w, "invalid token format", r.URL.Path)
			return
		}

		claims, err := auth.ParseToken(tokenString)
		if err != nil {
			model.Unauthorized(w, "invalid token", r.URL.Path)
			return
		}

		ctx := context.WithValue(r.Context(), "user_id", claims.UserID)
		next(w, r.WithContext(ctx))
	}
}
