package repository

import (
	"context"
	"errors"
	"log/slog"
	"toptal/internal/app/domain"
	"toptal/internal/app/repository/model"
	"toptal/internal/pkg/pg"
)

const (
	sqlFindUserByName = `SELECT * FROM users WHERE username = $1`
	sqlFindUserById   = `SELECT * FROM users WHERE id = $1`
	sqlCreateUser     = `INSERT INTO users (username, password_hash, admin) VALUES (:username, :password_hash, false)`
)

type UserRepository struct {
	db *pg.DB
}

func NewUserRepository(db *pg.DB) *UserRepository {
	return &UserRepository{db}
}

func (r *UserRepository) FindUserByName(ctx context.Context, name string) (domain.User, error) {
	var user model.User
	err := r.db.Get(ctx, "find_user_by_name", &user, sqlFindUserByName, name)
	if err != nil {
		slog.Error("failed to find user by name", "error", err, "name", name)
		return domain.User{}, errors.New("user not found")
	}
	return toDomainUser(user)
}

func (r *UserRepository) FindUserById(ctx context.Context, id int) (domain.User, error) {
	var user model.User
	err := r.db.Get(ctx, "find_user_by_id", &user, sqlFindUserById, id)
	if err != nil {
		slog.Error("failed to find user by id", "error", err, "id", id)
		return domain.User{}, errors.New("user not found")
	}
	return toDomainUser(user)
}

func (r *UserRepository) CreateUser(ctx context.Context, user domain.User) error {
	_, err := r.db.NamedExec(ctx, "create_user", sqlCreateUser, toModelUser(user))
	if err != nil {
		slog.Error("failed to insert user into database", "error", err)
		return errors.New("failed to create user")
	}
	return nil
}
