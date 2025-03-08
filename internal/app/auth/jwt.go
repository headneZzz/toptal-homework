package auth

import (
	"context"
	"errors"
	"log/slog"
	"net/http"
	"strings"
	"time"
	"toptal/internal/app/config"
	"toptal/internal/app/domain"

	"github.com/golang-jwt/jwt/v5"
)

var jwtConfig config.SecurityConfig

func SetConfig(cfg config.SecurityConfig) {
	jwtConfig = cfg
}

type Claims struct {
	UserID int `json:"user_id"`
	jwt.RegisteredClaims
}

func GenerateToken(user *domain.User) (string, error) {
	expirationTime := time.Now().Add(time.Duration(jwtConfig.JWTExpirationHours) * time.Hour)

	claims := &Claims{
		UserID: user.Id,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(expirationTime),
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString([]byte(jwtConfig.JWTSecret))
}

func JWTMiddleware(next func(w http.ResponseWriter, r *http.Request)) func(w http.ResponseWriter, r *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		authHeader := r.Header.Get("Authorization")
		if authHeader == "" {
			http.Error(w, "missing authorization header", http.StatusUnauthorized)
			return
		}

		tokenString := strings.TrimPrefix(authHeader, "Bearer ")
		if tokenString == "" {
			http.Error(w, "invalid token format", http.StatusUnauthorized)
			return
		}

		claims := &Claims{}
		token, err := jwt.ParseWithClaims(tokenString, claims, func(token *jwt.Token) (interface{}, error) {
			return []byte(jwtConfig.JWTSecret), nil
		})
		if err != nil || !token.Valid {
			http.Error(w, "invalid token", http.StatusUnauthorized)
			return
		}

		slog.Info("Token valid", "user", claims.UserID)
		ctx := context.WithValue(r.Context(), "user_id", claims.UserID)
		next(w, r.WithContext(ctx))
	}
}

func GetUserId(ctx context.Context) (int, error) {
	value := ctx.Value("user_id")
	if value == nil {
		return 0, errors.New("user not found in context")
	}
	return value.(int), nil
}
