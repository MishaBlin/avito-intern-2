package createReception

import (
	"avito-intern/internal/api/dto/internalErrors"
	"avito-intern/internal/api/dto/request/receptionDto"
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

func TestCreateReceptionHandler(t *testing.T) {
	tests := []struct {
		name           string
		requestData    receptionDto.CreateReceptionRequest
		userRole       string
		invalidBody    bool
		setupMock      func(mock *mockReceptionRepository)
		expectedStatus int
		expectedResp   interface{}
	}{
		{
			name: "Successful reception creation",
			requestData: receptionDto.CreateReceptionRequest{
				PVzID: "test-pvz-id",
			},
			userRole: "employee",
			setupMock: func(mockRepo *mockReceptionRepository) {
				mockRepo.On("GetActiveReception", "test-pvz-id").Return(nil, internalErrors.ErrNoActiveReception)
				mockRepo.On("CreateReception", mock.MatchedBy(func(reception *models.Reception) bool {
					return reception.PvzID == "test-pvz-id" && reception.Status == "in_progress"
				})).Return(nil)
			},
			expectedStatus: http.StatusCreated,
		},
		{
			name: "Active reception already exists",
			requestData: receptionDto.CreateReceptionRequest{
				PVzID: "test-pvz-id",
			},
			userRole: "employee",
			setupMock: func(mockRepo *mockReceptionRepository) {
				existingReception := &models.Reception{
					ID:     "existing-reception-id",
					PvzID:  "test-pvz-id",
					Status: "in_progress",
				}
				mockRepo.On("GetActiveReception", "test-pvz-id").Return(existingReception, nil)
			},
			expectedStatus: http.StatusBadRequest,
			expectedResp:   response.ErrorResponse{Message: "Active reception exists"},
		},
		{
			name: "Internal server error",
			requestData: receptionDto.CreateReceptionRequest{
				PVzID: "test-pvz-id",
			},
			userRole: "employee",
			setupMock: func(mockRepo *mockReceptionRepository) {
				mockRepo.On("GetActiveReception", "test-pvz-id").Return(nil, internalErrors.ErrNoActiveReception)
				mockRepo.On("CreateReception", mock.MatchedBy(func(reception *models.Reception) bool {
					return reception.PvzID == "test-pvz-id"
				})).Return(errors.New("database error"))
			},
			expectedStatus: http.StatusInternalServerError,
			expectedResp:   response.ErrorResponse{Message: "Internal server error"},
		},
		{
			name: "Access denied - moderator role",
			requestData: receptionDto.CreateReceptionRequest{
				PVzID: "test-pvz-id",
			},
			userRole:       "moderator",
			expectedStatus: http.StatusForbidden,
			expectedResp:   response.ErrorResponse{Message: "Access denied"},
		},
		{
			name:           "Invalid request body",
			invalidBody:    true,
			userRole:       "employee",
			expectedStatus: http.StatusBadRequest,
			expectedResp:   response.ErrorResponse{Message: "Invalid request"},
		},
		{
			name: "Missing PvzID in request",
			requestData: receptionDto.CreateReceptionRequest{
				PVzID: "",
			},
			userRole:       "employee",
			expectedStatus: http.StatusBadRequest,
			expectedResp:   response.ErrorResponse{Message: "Invalid request"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockRepo := new(mockReceptionRepository)
			if tt.setupMock != nil {
				tt.setupMock(mockRepo)
			}

			receptionService := services.NewReceptionService(mockRepo)

			handler := New(receptionService)

			var req *http.Request
			var err error

			if tt.invalidBody {
				req = httptest.NewRequest(http.MethodPost, "/reception", strings.NewReader("invalid json"))
			} else {
				body, _ := json.Marshal(tt.requestData)
				req = httptest.NewRequest(http.MethodPost, "/reception", bytes.NewReader(body))
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
				var receptionResp models.Reception
				err = json.Unmarshal(body, &receptionResp)
				require.NoError(t, err)
				require.Equal(t, tt.requestData.PVzID, receptionResp.PvzID)
				require.Equal(t, "in_progress", receptionResp.Status)
				require.NotEmpty(t, receptionResp.ID)
				require.False(t, receptionResp.DateTime.IsZero())
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
