package middleware

import (
	"net/http"
	"strings"
	"toptal/internal/app/auth"
	"toptal/internal/app/handler/model"
	"toptal/internal/app/util"
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

		ctx := util.WithUserID(r.Context(), claims.UserID)
		next(w, r.WithContext(ctx))
	}
}
