package service

import (
	"context"
	"errors"
	"fmt"
	"golang.org/x/crypto/bcrypt"
	"toptal/internal/app/auth"
	"toptal/internal/app/domain"
)

type UserService struct {
	userRepository UserRepository
}

func NewUserService(repository UserRepository) *UserService {
	return &UserService{repository}
}

func (s *UserService) Login(ctx context.Context, username string, password string) (string, error) {
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

func (s *UserService) CreateUser(ctx context.Context, username string, password string) error {
	hash, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return fmt.Errorf("failed to hash password: %w", err)
	}

	user := domain.User{Username: username, PasswordHash: string(hash)}

	return s.userRepository.CreateUser(ctx, user)
}

func (s *UserService) checkAdmin(ctx context.Context) error {
	userId, err := auth.GetUserId(ctx)
	if err != nil {
		return err
	}
	user, err := s.userRepository.FindUserById(ctx, userId)
	if err != nil {
		return err
	}
	if !user.Admin {
		return errors.New("user is not admin")
	}
	return nil
}
