package service

import (
	"context"
	"errors"
	"testing"

	"golang.org/x/crypto/bcrypt"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"

	"toptal/internal/app/domain"
	"toptal/internal/app/util"
)

type MockUserRepository struct {
	mock.Mock
}

func (m *MockUserRepository) FindUserByName(ctx context.Context, name string) (domain.User, error) {
	args := m.Called(ctx, name)
	return args.Get(0).(domain.User), args.Error(1)
}

func (m *MockUserRepository) FindUserById(ctx context.Context, id int) (domain.User, error) {
	args := m.Called(ctx, id)
	return args.Get(0).(domain.User), args.Error(1)
}

func (m *MockUserRepository) CreateUser(ctx context.Context, user domain.User) error {
	args := m.Called(ctx, user)
	return args.Error(0)
}

func TestAuthService_Login(t *testing.T) {
	ctx := context.Background()
	mockRepo := new(MockUserRepository)
	service := NewAuthService(mockRepo)

	t.Run("Successful login", func(t *testing.T) {
		// Create a user with known password hash
		password := "testpassword"
		hashedPassword, _ := HashPassword(password)
		user, err := domain.NewUser(1, "testuser", string(hashedPassword), false)
		assert.NoError(t, err)

		mockRepo.On("FindUserByName", ctx, "testuser").Return(user, nil)

		token, err := service.Login(ctx, "testuser", password)
		assert.NoError(t, err)
		assert.NotEmpty(t, token)

		mockRepo.AssertExpectations(t)
	})

	t.Run("User not found", func(t *testing.T) {
		mockRepo.On("FindUserByName", ctx, "nonexistent").
			Return(domain.User{}, errors.New("user not found"))

		token, err := service.Login(ctx, "nonexistent", "anypassword")
		assert.Error(t, err)
		assert.Empty(t, token)

		mockRepo.AssertExpectations(t)
	})

	t.Run("Invalid password", func(t *testing.T) {
		hashedPassword, _ := HashPassword("correctpassword")
		user, err := domain.NewUserWithDefaultId("testuser", string(hashedPassword))
		assert.NoError(t, err)

		mockRepo.On("FindUserByName", ctx, "testuser").Return(user, nil)

		token, err := service.Login(ctx, "testuser", "wrongpassword")
		assert.Error(t, err)
		assert.Empty(t, token)
		assert.Equal(t, "invalid password", err.Error())

		mockRepo.AssertExpectations(t)
	})
}

func TestAuthService_Register(t *testing.T) {
	ctx := context.Background()
	mockRepo := new(MockUserRepository)
	service := NewAuthService(mockRepo)

	t.Run("Successful registration", func(t *testing.T) {
		mockRepo.On("CreateUser", ctx, mock.MatchedBy(func(user domain.User) bool {
			return user.Username() == "newuser" && len(user.PasswordHash()) > 0
		})).Return(nil).Once()

		err := service.Register(ctx, "newuser", "password123")
		assert.NoError(t, err)

		mockRepo.AssertExpectations(t)
	})

	t.Run("Registration failure", func(t *testing.T) {
		mockRepo.On("CreateUser", ctx, mock.MatchedBy(func(user domain.User) bool {
			return user.Username() == "newuser" && len(user.PasswordHash()) > 0
		})).Return(errors.New("failed to create user")).Once()

		err := service.Register(ctx, "newuser", "password123")
		assert.Error(t, err)

		mockRepo.AssertExpectations(t)
	})
}

func TestAuthService_CheckAdmin(t *testing.T) {
	ctx := context.Background()
	mockRepo := new(MockUserRepository)
	service := NewAuthService(mockRepo)

	t.Run("User is admin", func(t *testing.T) {
		// Create context with user ID
		ctx = util.WithUserID(ctx, 1)
		user, err := domain.NewUser(1, "adminuser", "password", true)
		assert.NoError(t, err)

		mockRepo.On("FindUserById", ctx, 1).Return(user, nil)

		err = service.checkAdmin(ctx)
		assert.NoError(t, err)

		mockRepo.AssertExpectations(t)
	})

	t.Run("User is not admin", func(t *testing.T) {
		ctx = util.WithUserID(ctx, 2)
		user, err := domain.NewUser(2, "testuser", "password", false)
		assert.NoError(t, err)

		mockRepo.On("FindUserById", ctx, 2).Return(user, nil)

		err = service.checkAdmin(ctx)
		assert.Error(t, err)
		assert.Equal(t, "forbidden", err.Error())

		mockRepo.AssertExpectations(t)
	})

	t.Run("User not found", func(t *testing.T) {
		ctx = util.WithUserID(ctx, 3)

		mockRepo.On("FindUserById", ctx, 3).Return(domain.User{}, errors.New("user not found"))

		err := service.checkAdmin(ctx)
		assert.Error(t, err)

		mockRepo.AssertExpectations(t)
	})

	t.Run("No user ID in context", func(t *testing.T) {
		err := service.checkAdmin(context.Background())
		assert.Error(t, err)
	})
}

// Helper function to create a context with user ID
func TestContextWithUserId(t *testing.T) {
	ctx := context.Background()
	userId := 1

	ctxWithUser := util.WithUserID(ctx, userId)
	assert.NotNil(t, ctxWithUser)

	extractedId, err := util.GetUserID(ctxWithUser)
	assert.NoError(t, err)
	assert.Equal(t, userId, extractedId)
}

func HashPassword(password string) ([]byte, error) {
	return bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
}
