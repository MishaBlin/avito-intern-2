package services

import (
	"avito-intern/internal/api/dto/internalErrors"
	"avito-intern/internal/api/dto/request/authDto"
	"avito-intern/internal/models"
	"avito-intern/internal/repository"
	"avito-intern/internal/utils"
	"errors"

	"github.com/google/uuid"
)

type AuthService struct {
	userRepo repository.UserRepositoryInterface
}

func NewAuthService(userRepo repository.UserRepositoryInterface) *AuthService {
	return &AuthService{
		userRepo: userRepo,
	}
}

func (s *AuthService) RegisterUser(req authDto.RegisterRequest) (*models.User, error) {
	_, err := s.getUserByEmail(req.Email)

	if !errors.Is(err, internalErrors.ErrUserNotFound) {
		if err == nil {
			return nil, internalErrors.ErrEmailExists
		}
		return nil, err
	}

	user := &models.User{
		ID:       uuid.New().String(),
		Email:    req.Email,
		Role:     req.Role,
		Password: utils.HashPassword(req.Password),
	}

	err = s.userRepo.CreateUser(user)
	return user, err
}

func (s *AuthService) AuthenticateUser(req authDto.LoginRequest) (string, error) {
	user, err := s.validateCredentials(req.Email, req.Password)
	if err != nil {
		return "", err
	}
	return utils.GenerateJWT(user.ID, user.Role)
}

func (s *AuthService) getUserByEmail(email string) (*models.User, error) {
	return s.userRepo.GetUserByEmail(email)
}

func (s *AuthService) validateCredentials(email, password string) (*models.User, error) {
	user, err := s.getUserByEmail(email)
	if err != nil {
		return nil, internalErrors.ErrInvalidCredentials
	}
	hashedPassword := utils.HashPassword(password)
	if user.Password != hashedPassword {
		return nil, internalErrors.ErrInvalidCredentials
	}
	return user, nil
}
