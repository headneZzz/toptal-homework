package service

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"toptal/internal/app/auth"
	"toptal/internal/app/domain"
	"toptal/internal/app/util"

	"golang.org/x/crypto/bcrypt"
)

type AuthService struct {
	userRepository UserRepository
}

func NewAuthService(repository UserRepository) *AuthService {
	return &AuthService{repository}
}

func (s *AuthService) Login(ctx context.Context, username string, password string) (string, error) {
	user, err := s.userRepository.FindUserByName(ctx, username)
	if err != nil {
		return "", err
	}

	if err := bcrypt.CompareHashAndPassword([]byte(user.PasswordHash), []byte(password)); err != nil {
		return "", errors.New("invalid password")
	}

	token, err := auth.GenerateToken(&user)
	if err != nil {
		return "", errors.New("failed to generate token")
	}

	return token, nil
}

func (s *AuthService) Register(ctx context.Context, username string, password string) error {
	hash, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return fmt.Errorf("failed to hash password: %w", err)
	}

	user := domain.User{Username: username, PasswordHash: string(hash)}

	return s.userRepository.CreateUser(ctx, user)
}

func (s *AuthService) checkAdmin(ctx context.Context) error {
	userId, err := util.GetUserID(ctx)
	if err != nil {
		slog.Error("failed to get user ID from context", "error", err)
		return errors.New("failed to get user ID from context")
	}
	user, err := s.userRepository.FindUserById(ctx, userId)
	if err != nil {
		slog.Error("failed to find user by ID", "userId", userId, "error", err)
		return errors.New("failed to find user by ID")
	}
	if !user.Admin {
		return domain.ErrForbidden
	}
	return nil
}
