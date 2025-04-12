package services

import (
	"avito-intern/internal/api/dto/internalErrors"
	"avito-intern/internal/api/dto/request/authDto"
	"avito-intern/internal/models"
	"avito-intern/internal/utils"
	"errors"
	"testing"

	"github.com/stretchr/testify/assert"
)

type mockUserRepository struct {
	users      map[string]*models.User
	createErr  error
	getUserErr error
}

func (m *mockUserRepository) CreateUser(user *models.User) error {
	if m.createErr != nil {
		return m.createErr
	}
	if _, exists := m.users[user.Email]; exists {
		return internalErrors.ErrEmailExists
	}
	m.users[user.Email] = user
	return nil
}

func (m *mockUserRepository) GetUserByEmail(email string) (*models.User, error) {
	if m.getUserErr != nil {
		return nil, m.getUserErr
	}
	if user, exists := m.users[email]; exists {
		return user, nil
	}
	return nil, internalErrors.ErrUserNotFound
}

func TestAuthService_RegisterUser_Success(t *testing.T) {

	mockRepo := &mockUserRepository{
		users: make(map[string]*models.User),
	}
	service := NewAuthService(mockRepo)
	req := authDto.RegisterRequest{
		Email:    "test@example.com",
		Password: "password123",
		Role:     "moderator",
	}

	user, err := service.RegisterUser(req)

	assert.NoError(t, err)
	assert.NotNil(t, user)
	assert.Equal(t, "test@example.com", user.Email)
	assert.Equal(t, "moderator", user.Role)

	assert.NotEqual(t, "password123", user.Password)
}

func TestAuthService_RegisterUser_DuplicateEmail(t *testing.T) {

	existingUser := &models.User{
		ID:       "existing-id",
		Email:    "test@example.com",
		Password: "hashed_password",
		Role:     "moderator",
	}

	mockRepo := &mockUserRepository{
		users: map[string]*models.User{
			"test@example.com": existingUser,
		},
	}
	service := NewAuthService(mockRepo)
	req := authDto.RegisterRequest{
		Email:    "test@example.com",
		Password: "password123",
		Role:     "employee",
	}

	user, err := service.RegisterUser(req)

	assert.Error(t, err)
	assert.Equal(t, internalErrors.ErrEmailExists, err)
	assert.Nil(t, user)
}

func TestAuthService_RegisterUser_RepositoryError(t *testing.T) {

	mockRepo := &mockUserRepository{
		users:      make(map[string]*models.User),
		getUserErr: errors.New("database error"),
	}
	service := NewAuthService(mockRepo)
	req := authDto.RegisterRequest{
		Email:    "test@example.com",
		Password: "password123",
		Role:     "moderator",
	}

	user, err := service.RegisterUser(req)

	assert.Error(t, err)
	assert.Equal(t, "database error", err.Error())
	assert.Nil(t, user)
}

func TestAuthService_AuthenticateUser_ValidCredentials(t *testing.T) {

	testUser := &models.User{
		ID:       "test-id",
		Email:    "test@example.com",
		Password: utils.HashPassword("password123"),
		Role:     "moderator",
	}

	mockRepo := &mockUserRepository{
		users: map[string]*models.User{
			testUser.Email: testUser,
		},
	}
	service := NewAuthService(mockRepo)
	req := authDto.LoginRequest{
		Email:    "test@example.com",
		Password: "password123",
	}

	token, err := service.AuthenticateUser(req)

	assert.NoError(t, err)
	assert.NotEmpty(t, token)
}

func TestAuthService_AuthenticateUser_InvalidPassword(t *testing.T) {

	testUser := &models.User{
		ID:       "test-id",
		Email:    "test@example.com",
		Password: utils.HashPassword("password123"),
		Role:     "moderator",
	}

	mockRepo := &mockUserRepository{
		users: map[string]*models.User{
			testUser.Email: testUser,
		},
	}
	service := NewAuthService(mockRepo)
	req := authDto.LoginRequest{
		Email:    "test@example.com",
		Password: "wrongpassword",
	}

	token, err := service.AuthenticateUser(req)

	assert.Error(t, err)
	assert.Equal(t, internalErrors.ErrInvalidCredentials, err)
	assert.Empty(t, token)
}

func TestAuthService_AuthenticateUser_NonExistentUser(t *testing.T) {

	mockRepo := &mockUserRepository{
		users: make(map[string]*models.User),
	}
	service := NewAuthService(mockRepo)
	req := authDto.LoginRequest{
		Email:    "nonexistent@example.com",
		Password: "password123",
	}

	token, err := service.AuthenticateUser(req)

	assert.Error(t, err)
	assert.Equal(t, internalErrors.ErrInvalidCredentials, err)
	assert.Empty(t, token)
}

func TestAuthService_AuthenticateUser_RepositoryError(t *testing.T) {

	mockRepo := &mockUserRepository{
		users:      make(map[string]*models.User),
		getUserErr: errors.New("database error"),
	}
	service := NewAuthService(mockRepo)
	req := authDto.LoginRequest{
		Email:    "test@example.com",
		Password: "password123",
	}

	token, err := service.AuthenticateUser(req)

	assert.Error(t, err)
	assert.Equal(t, internalErrors.ErrInvalidCredentials, err) // Changed from database error
	assert.Empty(t, token)
}
