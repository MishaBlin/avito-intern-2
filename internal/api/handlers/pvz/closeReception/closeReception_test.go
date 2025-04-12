package closeReception

import (
	"avito-intern/internal/api/dto/internalErrors"
	"avito-intern/internal/api/dto/response"
	"avito-intern/internal/api/middleware"
	"avito-intern/internal/models"
	"avito-intern/internal/services"
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

type mockReceptionRepository struct {
	mock.Mock
}

func (m *mockReceptionRepository) CreateReception(reception *models.Reception) error {
	args := m.Called(reception)
	return args.Error(0)
}

func (m *mockReceptionRepository) GetActiveReception(pvzID string) (*models.Reception, error) {
	args := m.Called(pvzID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*models.Reception), args.Error(1)
}

func (m *mockReceptionRepository) CloseReception(receptionID string) error {
	args := m.Called(receptionID)
	return args.Error(0)
}

func createUserContext(role string) context.Context {
	user := models.User{
		ID:    uuid.New().String(),
		Email: "test@example.com",
		Role:  role,
	}
	return context.WithValue(context.Background(), middleware.UserCtxKey, user)
}

func TestCloseReceptionHandler(t *testing.T) {
	tests := []struct {
		name           string
		pvzID          string
		userRole       string
		setupMock      func(mockReceptionRepo *mockReceptionRepository)
		expectedStatus int
		expectedResp   interface{}
	}{
		{
			name:     "Successful close",
			pvzID:    "test-pvz",
			userRole: "employee",
			setupMock: func(mockReceptionRepo *mockReceptionRepository) {
				reception := &models.Reception{
					ID:       "reception-id",
					PvzID:    "test-pvz",
					DateTime: time.Now(),
					Status:   "active",
				}

				mockReceptionRepo.On("GetActiveReception", "test-pvz").Return(reception, nil)
				mockReceptionRepo.On("CloseReception", "reception-id").Return(nil)
			},
			expectedStatus: http.StatusOK,
			expectedResp: &models.Reception{
				ID:     "reception-id",
				PvzID:  "test-pvz",
				Status: "close",
			},
		},
		{
			name:     "No active reception",
			pvzID:    "test-pvz",
			userRole: "employee",
			setupMock: func(mockReceptionRepo *mockReceptionRepository) {
				mockReceptionRepo.On("GetActiveReception", "test-pvz").Return(nil, internalErrors.ErrNoActiveReception)
			},
			expectedStatus: http.StatusBadRequest,
			expectedResp:   response.ErrorResponse{Message: "No active reception"},
		},
		{
			name:     "Internal server error",
			pvzID:    "test-pvz",
			userRole: "employee",
			setupMock: func(mockReceptionRepo *mockReceptionRepository) {
				reception := &models.Reception{
					ID:       "reception-id",
					PvzID:    "test-pvz",
					DateTime: time.Now(),
					Status:   "active",
				}

				mockReceptionRepo.On("GetActiveReception", "test-pvz").Return(reception, nil)
				mockReceptionRepo.On("CloseReception", "reception-id").Return(errors.New("database error"))
			},
			expectedStatus: http.StatusInternalServerError,
			expectedResp:   response.ErrorResponse{Message: "Internal server error"},
		},
		{
			name:           "Access denied",
			pvzID:          "test-pvz",
			userRole:       "user",
			expectedStatus: http.StatusForbidden,
			expectedResp:   response.ErrorResponse{Message: "Access denied"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockReceptionRepo := new(mockReceptionRepository)

			if tt.setupMock != nil {
				tt.setupMock(mockReceptionRepo)
			}

			receptionService := services.NewReceptionService(mockReceptionRepo)

			r := chi.NewRouter()
			r.Post("/pvz/{pvzId}/close_reception", New(receptionService))

			req := httptest.NewRequest(http.MethodPost, "/pvz/"+tt.pvzID+"/close_reception", nil)

			ctx := createUserContext(tt.userRole)
			req = req.WithContext(ctx)

			w := httptest.NewRecorder()
			r.ServeHTTP(w, req)

			require.Equal(t, tt.expectedStatus, w.Code)

			if tt.expectedStatus == http.StatusOK {
				var reception models.Reception
				err := json.NewDecoder(w.Body).Decode(&reception)
				require.NoError(t, err)
				require.Equal(t, tt.expectedResp.(*models.Reception).ID, reception.ID)
				require.Equal(t, tt.expectedResp.(*models.Reception).PvzID, reception.PvzID)
				require.Equal(t, tt.expectedResp.(*models.Reception).Status, reception.Status)
			} else {
				var errorResp response.ErrorResponse
				err := json.NewDecoder(w.Body).Decode(&errorResp)
				require.NoError(t, err)
				require.Equal(t, tt.expectedResp, errorResp)
			}

			mockReceptionRepo.AssertExpectations(t)
		})
	}
}
