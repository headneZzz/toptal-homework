package service

import (
	"context"
	"errors"
	"fmt"
	"golang.org/x/crypto/bcrypt"
	"toptal/internal/app/auth"
	"toptal/internal/app/domain"
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

	if err := bcrypt.CompareHashAndPassword([]byte(user.PasswordHash()), []byte(password)); err != nil {
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

	var user domain.User
	if err = user.SetUsername(username); err != nil {
		return fmt.Errorf("failed to set username: %w", err)
	}
	if err = user.SetPasswordHash(string(hash)); err != nil {
		return fmt.Errorf("failed to set password hash: %w", err)
	}

	return s.userRepository.CreateUser(ctx, user)
}

func (s *AuthService) GetUserById(ctx context.Context, id int) (domain.User, error) {
	return s.userRepository.FindUserById(ctx, id)
}
