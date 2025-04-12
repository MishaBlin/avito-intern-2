package register

import (
	"avito-intern/internal/api/dto/internalErrors"
	"avito-intern/internal/api/dto/request/authDto"
	"avito-intern/internal/api/dto/response"
	"avito-intern/internal/models"
	"avito-intern/internal/services"
	"bytes"
	"encoding/json"
	"errors"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

type mockUserRepository struct {
	mock.Mock
}

func (m *mockUserRepository) CreateUser(user *models.User) error {
	args := m.Called(user)
	return args.Error(0)
}

func (m *mockUserRepository) GetUserByEmail(email string) (*models.User, error) {
	args := m.Called(email)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*models.User), args.Error(1)
}

func createMockAuthService(t *testing.T) (*services.AuthService, *mockUserRepository) {
	mockRepo := new(mockUserRepository)
	authService := services.NewAuthService(mockRepo)
	return authService, mockRepo
}

func TestRegisterHandler(t *testing.T) {
	tests := []struct {
		name           string
		registerData   authDto.RegisterRequest
		invalidBody    bool
		setupMockRepo  func(mockRepo *mockUserRepository)
		expectedStatus int
		expectedResp   interface{}
	}{
		{
			name: "Successful registration",
			registerData: authDto.RegisterRequest{
				Email:    "test@example.com",
				Password: "password123",
				Role:     "employee",
			},
			setupMockRepo: func(mockRepo *mockUserRepository) {
				mockRepo.On("GetUserByEmail", "test@example.com").Return(nil, internalErrors.ErrUserNotFound)
				mockRepo.On("CreateUser", mock.AnythingOfType("*models.User")).Return(nil)
			},
			expectedStatus: http.StatusCreated,
			expectedResp: response.UserResponse{
				Email: "test@example.com",
				Role:  "employee",
			},
		},
		{
			name: "Email already exists",
			registerData: authDto.RegisterRequest{
				Email:    "existing@example.com",
				Password: "password123",
				Role:     "employee",
			},
			setupMockRepo: func(mockRepo *mockUserRepository) {
				// Return an existing user
				mockRepo.On("GetUserByEmail", "existing@example.com").Return(&models.User{
					ID:    "existing-id",
					Email: "existing@example.com",
				}, nil)
			},
			expectedStatus: http.StatusBadRequest,
			expectedResp:   response.ErrorResponse{Message: "Email already exists"},
		},
		{
			name: "Invalid role",
			registerData: authDto.RegisterRequest{
				Email:    "test@example.com",
				Password: "password123",
				Role:     "invalid-role",
			},
			expectedStatus: http.StatusBadRequest,
			expectedResp:   response.ErrorResponse{Message: "Invalid role"},
		},
		{
			name: "Internal server error",
			registerData: authDto.RegisterRequest{
				Email:    "test@example.com",
				Password: "password123",
				Role:     "moderator",
			},
			setupMockRepo: func(mockRepo *mockUserRepository) {
				mockRepo.On("GetUserByEmail", "test@example.com").Return(nil, internalErrors.ErrUserNotFound)
				mockRepo.On("CreateUser", mock.AnythingOfType("*models.User")).Return(errors.New("database error"))
			},
			expectedStatus: http.StatusInternalServerError,
			expectedResp:   response.ErrorResponse{Message: "Error registering user"},
		},
		{
			name:           "Invalid request body",
			invalidBody:    true,
			expectedStatus: http.StatusBadRequest,
			expectedResp:   response.ErrorResponse{Message: "Invalid request"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			authService, mockRepo := createMockAuthService(t)

			if tt.setupMockRepo != nil {
				tt.setupMockRepo(mockRepo)
			}

			handler := New(authService)

			var req *http.Request
			var err error

			if tt.invalidBody {
				req = httptest.NewRequest(http.MethodPost, "/register", strings.NewReader("invalid json"))
			} else {
				body, _ := json.Marshal(tt.registerData)
				req = httptest.NewRequest(http.MethodPost, "/register", bytes.NewReader(body))
			}

			req.Header.Set("Content-Type", "application/json")
			w := httptest.NewRecorder()

			handler.ServeHTTP(w, req)

			require.Equal(t, tt.expectedStatus, w.Code)

			var resp interface{}
			if tt.invalidBody {
				body, err := io.ReadAll(w.Body)
				require.NoError(t, err)

				var errorResp response.ErrorResponse
				err = json.Unmarshal(body, &errorResp)
				require.NoError(t, err)
				resp = errorResp
			} else {
				if tt.expectedStatus == http.StatusCreated {
					var userResp response.UserResponse
					err = json.NewDecoder(w.Body).Decode(&userResp)
					resp = userResp

					require.Equal(t, tt.expectedResp.(response.UserResponse).Email, userResp.Email)
					require.Equal(t, tt.expectedResp.(response.UserResponse).Role, userResp.Role)
					require.NotEmpty(t, userResp.ID)
					return
				} else {
					var errorResp response.ErrorResponse
					err = json.NewDecoder(w.Body).Decode(&errorResp)
					resp = errorResp
				}
				require.NoError(t, err)
			}

			require.Equal(t, tt.expectedResp, resp)

			mockRepo.AssertExpectations(t)
		})
	}
}
