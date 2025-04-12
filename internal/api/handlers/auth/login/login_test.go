package login

import (
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

	"avito-intern/internal/api/dto/internalErrors"
	"avito-intern/internal/api/dto/request/authDto"
	"avito-intern/internal/api/dto/response"
	"avito-intern/internal/models"
)

type mockAuthService struct {
	mock.Mock
}

var _ AuthService = (*mockAuthService)(nil)

func (m *mockAuthService) RegisterUser(req authDto.RegisterRequest) (*models.User, error) {
	args := m.Called(req)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*models.User), args.Error(1)
}

func (m *mockAuthService) AuthenticateUser(req authDto.LoginRequest) (string, error) {
	args := m.Called(req)
	return args.String(0), args.Error(1)
}

func TestLoginHandler(t *testing.T) {
	tests := []struct {
		name           string
		credentials    authDto.LoginRequest
		invalidBody    bool
		setupMock      func(mock *mockAuthService)
		expectedStatus int
		expectedResp   interface{}
	}{
		{
			name: "Successful login",
			credentials: authDto.LoginRequest{
				Email:    "test@example.com",
				Password: "password123",
			},
			setupMock: func(mock *mockAuthService) {
				mock.On("AuthenticateUser", authDto.LoginRequest{
					Email:    "test@example.com",
					Password: "password123",
				}).Return("token123", nil)
			},
			expectedStatus: http.StatusOK,
			expectedResp:   response.TokenResponse{Token: "token123"},
		},
		{
			name: "Invalid credentials",
			credentials: authDto.LoginRequest{
				Email:    "test@example.com",
				Password: "wrongpass",
			},
			setupMock: func(mock *mockAuthService) {
				mock.On("AuthenticateUser", authDto.LoginRequest{
					Email:    "test@example.com",
					Password: "wrongpass",
				}).Return("", internalErrors.ErrInvalidCredentials)
			},
			expectedStatus: http.StatusUnauthorized,
			expectedResp:   response.ErrorResponse{Message: "Invalid credentials"},
		},
		{
			name: "Internal server error",
			credentials: authDto.LoginRequest{
				Email:    "test@example.com",
				Password: "password123",
			},
			setupMock: func(mock *mockAuthService) {
				mock.On("AuthenticateUser", authDto.LoginRequest{
					Email:    "test@example.com",
					Password: "password123",
				}).Return("", errors.New("internal error"))
			},
			expectedStatus: http.StatusInternalServerError,
			expectedResp:   response.ErrorResponse{Message: "Could not generate token"},
		},
		{
			name:           "Invalid request body",
			invalidBody:    true,
			expectedStatus: http.StatusBadRequest,
			expectedResp:   "Invalid request",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockService := new(mockAuthService)
			if tt.setupMock != nil {
				tt.setupMock(mockService)
			}

			handler := New(mockService)

			var req *http.Request
			var err error

			if tt.invalidBody {
				req = httptest.NewRequest(http.MethodPost, "/login", strings.NewReader("invalid json"))
			} else {
				body, _ := json.Marshal(tt.credentials)
				req = httptest.NewRequest(http.MethodPost, "/login", bytes.NewReader(body))
			}

			req.Header.Set("Content-Type", "application/json")
			w := httptest.NewRecorder()

			handler.ServeHTTP(w, req)

			require.Equal(t, tt.expectedStatus, w.Code)

			if tt.invalidBody {
				body, err := io.ReadAll(w.Body)
				require.NoError(t, err)
				require.Equal(t, tt.expectedResp, strings.TrimSpace(string(body)))
			} else {
				var resp interface{}
				if tt.expectedStatus == http.StatusOK {
					var successResp response.TokenResponse
					err = json.NewDecoder(w.Body).Decode(&successResp)
					resp = successResp
				} else {
					var errorResp response.ErrorResponse
					err = json.NewDecoder(w.Body).Decode(&errorResp)
					resp = errorResp
				}
				require.NoError(t, err)
				require.Equal(t, tt.expectedResp, resp)
			}

			mockService.AssertExpectations(t)
		})
	}
}
