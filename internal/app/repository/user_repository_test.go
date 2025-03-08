package repository

import (
	"context"
	"database/sql"
	"errors"
	"testing"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/jmoiron/sqlx"
	"github.com/stretchr/testify/assert"
	"toptal/internal/app/domain"
	"toptal/internal/pkg/pg"
)

func TestUserRepository_FindUserByName(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("an error '%s' was not expected when opening a stub database connection", err)
	}
	defer db.Close()

	pgDB := pg.NewDB(sqlx.NewDb(db, "sqlmock"))
	repo := NewUserRepository(pgDB)

	// Тест: пользователь найден
	t.Run("User found", func(t *testing.T) {
		rows := sqlmock.NewRows([]string{"id", "username", "password_hash", "admin"}).
			AddRow(1, "testuser", "hash", false)

		mock.ExpectQuery("SELECT \\* FROM users WHERE username = \\$1").
			WithArgs("testuser").
			WillReturnRows(rows)

		user, err := repo.FindUserByName(context.Background(), "testuser")
		assert.NoError(t, err)
		assert.Equal(t, "testuser", user.Username)
	})

	// Тест: пользователь не найден
	t.Run("User not found", func(t *testing.T) {
		mock.ExpectQuery("SELECT \\* FROM users WHERE username = \\$1").
			WithArgs("unknown").
			WillReturnError(sql.ErrNoRows)

		_, err := repo.FindUserByName(context.Background(), "unknown")
		assert.Error(t, err)
		assert.Equal(t, "user not found", err.Error())
	})
}

func TestUserRepository_FindUserById(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("an error '%s' was not expected when opening a stub database connection", err)
	}
	defer db.Close()

	pgDB := pg.NewDB(sqlx.NewDb(db, "sqlmock"))
	repo := NewUserRepository(pgDB)

	// Тест: пользователь найден
	t.Run("User found", func(t *testing.T) {
		rows := sqlmock.NewRows([]string{"id", "username", "password_hash", "admin"}).
			AddRow(1, "testuser", "hash", false)

		mock.ExpectQuery("SELECT \\* FROM users WHERE id = \\$1").
			WithArgs(1).
			WillReturnRows(rows)

		user, err := repo.FindUserById(context.Background(), 1)
		assert.NoError(t, err)
		assert.Equal(t, "testuser", user.Username)
	})

	// Тест: пользователь не найден
	t.Run("User not found", func(t *testing.T) {
		mock.ExpectQuery("SELECT \\* FROM users WHERE id = \\$1").
			WithArgs(999).
			WillReturnError(sql.ErrNoRows)

		_, err := repo.FindUserById(context.Background(), 999)
		assert.Error(t, err)
		assert.Equal(t, "user not found", err.Error())
	})
}

func TestUserRepository_CreateUser(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("an error '%s' was not expected when opening a stub database connection", err)
	}
	defer db.Close()

	sqlxDB := sqlx.NewDb(db, "sqlmock")
	repo := NewUserRepository(pg.NewDB(sqlxDB))

	// Тест: пользователь успешно создан
	t.Run("User created", func(t *testing.T) {
		user := domain.User{
			Username:     "testuser",
			PasswordHash: "hash",
		}

		mock.ExpectExec("INSERT INTO users").
			WithArgs(user.Username, user.PasswordHash).
			WillReturnResult(sqlmock.NewResult(1, 1))

		err := repo.CreateUser(context.Background(), user)
		assert.NoError(t, err)
	})

	// Тест: ошибка при создании пользователя
	t.Run("Create user error", func(t *testing.T) {
		user := domain.User{
			Username:     "testuser",
			PasswordHash: "hash",
		}

		mock.ExpectExec("INSERT INTO users").
			WithArgs(user.Username, user.PasswordHash, false).
			WillReturnError(errors.New("duplicate key"))

		err := repo.CreateUser(context.Background(), user)
		assert.Error(t, err)
		assert.Equal(t, "failed to create user", err.Error())
	})
}
