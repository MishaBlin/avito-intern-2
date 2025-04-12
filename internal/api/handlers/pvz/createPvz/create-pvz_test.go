package createPvz

import (
	"avito-intern/internal/api/dto/response"
	"avito-intern/internal/api/middleware"
	"avito-intern/internal/models"
	"avito-intern/internal/services"
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

type mockPVZRepository struct {
	mock.Mock
}

func (m *mockPVZRepository) CreatePVZ(pvz *models.PVZ) error {
	args := m.Called(pvz)
	return args.Error(0)
}

func (m *mockPVZRepository) ListPVZ(limit, offset int, startDate, endDate *time.Time) ([]*models.PVZ, error) {
	args := m.Called(limit, offset, startDate, endDate)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]*models.PVZ), args.Error(1)
}

func createUserContext(role string) context.Context {
	user := models.User{
		ID:    uuid.New().String(),
		Email: "test@example.com",
		Role:  role,
	}
	return context.WithValue(context.Background(), middleware.UserCtxKey, user)
}

func TestCreatePVZHandler(t *testing.T) {
	tests := []struct {
		name           string
		pvzData        models.PVZ
		userRole       string
		invalidBody    bool
		setupMock      func(mock *mockPVZRepository)
		expectedStatus int
		expectedResp   interface{}
	}{
		{
			name: "Successful PVZ creation",
			pvzData: models.PVZ{
				ID:   "test-pvz-id",
				City: "Москва",
			},
			userRole: "moderator",
			setupMock: func(mockRepo *mockPVZRepository) {
				mockRepo.On("CreatePVZ", mock.MatchedBy(func(pvz *models.PVZ) bool {
					return pvz.City == "Москва" && pvz.ID == "test-pvz-id"
				})).Return(nil)
			},
			expectedStatus: http.StatusCreated,
		},
		{
			name: "Invalid city",
			pvzData: models.PVZ{
				ID:   "test-pvz-id",
				City: "Invalid City",
			},
			userRole:       "moderator",
			expectedStatus: http.StatusBadRequest,
			expectedResp:   response.ErrorResponse{Message: "City not allowed"},
		},
		{
			name: "Internal server error",
			pvzData: models.PVZ{
				ID:   "test-pvz-id",
				City: "Москва",
			},
			userRole: "moderator",
			setupMock: func(mockRepo *mockPVZRepository) {
				mockRepo.On("CreatePVZ", mock.MatchedBy(func(pvz *models.PVZ) bool {
					return pvz.City == "Москва"
				})).Return(errors.New("database error"))
			},
			expectedStatus: http.StatusInternalServerError,
			expectedResp:   response.ErrorResponse{Message: "Internal server error"},
		},
		{
			name: "Access denied - employee role",
			pvzData: models.PVZ{
				ID:   "test-pvz-id",
				City: "Москва",
			},
			userRole:       "employee",
			expectedStatus: http.StatusForbidden,
			expectedResp:   response.ErrorResponse{Message: "Access denied"},
		},
		{
			name:           "Invalid request body",
			invalidBody:    true,
			userRole:       "moderator",
			expectedStatus: http.StatusBadRequest,
			expectedResp:   response.ErrorResponse{Message: "Invalid request"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockRepo := new(mockPVZRepository)
			if tt.setupMock != nil {
				tt.setupMock(mockRepo)
			}

			pvzService := services.NewPVZService(mockRepo)

			handler := New(pvzService)

			var req *http.Request
			var err error

			if tt.invalidBody {
				req = httptest.NewRequest(http.MethodPost, "/pvz", strings.NewReader("invalid json"))
			} else {
				body, _ := json.Marshal(tt.pvzData)
				req = httptest.NewRequest(http.MethodPost, "/pvz", bytes.NewReader(body))
			}

			ctx := createUserContext(tt.userRole)
			req = req.WithContext(ctx)

			req.Header.Set("Content-Type", "application/json")
			w := httptest.NewRecorder()

			handler.ServeHTTP(w, req)

			require.Equal(t, tt.expectedStatus, w.Code)

			body, err := io.ReadAll(w.Body)
			require.NoError(t, err)

			if tt.expectedStatus == http.StatusCreated {
				var pvzResp models.PVZ
				err = json.Unmarshal(body, &pvzResp)
				require.NoError(t, err)
				require.Equal(t, tt.pvzData.ID, pvzResp.ID)
				require.Equal(t, tt.pvzData.City, pvzResp.City)
			} else {
				var errorResp response.ErrorResponse
				err = json.Unmarshal(body, &errorResp)
				require.NoError(t, err)
				require.Equal(t, tt.expectedResp, errorResp)
			}

			mockRepo.AssertExpectations(t)
		})
	}
}
