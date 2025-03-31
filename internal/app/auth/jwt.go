package auth

import (
	"errors"
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
		UserID: user.Id(),
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(expirationTime),
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString([]byte(jwtConfig.JWTSecret))
}

func ParseToken(tokenString string) (*Claims, error) {
	claims := &Claims{}
	token, err := jwt.ParseWithClaims(tokenString, claims, func(token *jwt.Token) (interface{}, error) {
		return []byte(jwtConfig.JWTSecret), nil
	})

	if err != nil || !token.Valid {
		return nil, errors.New("invalid token")
	}
	return claims, nil
}
