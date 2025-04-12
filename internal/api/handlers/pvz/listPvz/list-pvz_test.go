package listPvz

import (
	"avito-intern/internal/api/dto/response"
	"avito-intern/internal/api/middleware"
	"avito-intern/internal/models"
	"avito-intern/internal/services"
	"context"
	"encoding/json"
	"errors"
	"io"
	"net/http"
	"net/http/httptest"
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

func TestListPVZHandler(t *testing.T) {
	mockPVZData := []*models.PVZ{
		{
			ID:               "pvz-1",
			City:             "Москва",
			RegistrationDate: time.Date(2023, 1, 1, 0, 0, 0, 0, time.UTC),
		},
		{
			ID:               "pvz-2",
			City:             "Санкт-Петербург",
			RegistrationDate: time.Date(2023, 2, 1, 0, 0, 0, 0, time.UTC),
		},
	}

	tests := []struct {
		name           string
		userRole       string
		queryParams    map[string]string
		setupMock      func(mock *mockPVZRepository)
		expectedStatus int
		expectedResp   interface{}
	}{
		{
			name:     "Successfully list PVZs - employee role",
			userRole: "employee",
			queryParams: map[string]string{
				"limit": "10",
				"page":  "1",
			},
			setupMock: func(mockRepo *mockPVZRepository) {
				mockRepo.On("ListPVZ", 10, 0, (*time.Time)(nil), (*time.Time)(nil)).Return(mockPVZData, nil)
			},
			expectedStatus: http.StatusOK,
			expectedResp:   mockPVZData,
		},
		{
			name:     "Successfully list PVZs - moderator role",
			userRole: "moderator",
			queryParams: map[string]string{
				"limit": "5",
				"page":  "2",
			},
			setupMock: func(mockRepo *mockPVZRepository) {
				mockRepo.On("ListPVZ", 5, 5, (*time.Time)(nil), (*time.Time)(nil)).Return(mockPVZData, nil)
			},
			expectedStatus: http.StatusOK,
			expectedResp:   mockPVZData,
		},
		{
			name:     "Successfully list PVZs with date filters",
			userRole: "employee",
			queryParams: map[string]string{
				"limit":     "10",
				"page":      "1",
				"startDate": "2023-01-01T00:00:00",
				"endDate":   "2023-12-31T23:59:59",
			},
			setupMock: func(mockRepo *mockPVZRepository) {
				startDate, _ := time.Parse("2006-01-02T15:04:05", "2023-01-01T00:00:00")
				endDate, _ := time.Parse("2006-01-02T15:04:05", "2023-12-31T23:59:59")
				mockRepo.On("ListPVZ", 10, 0, mock.MatchedBy(func(t *time.Time) bool {
					return t != nil && t.Equal(startDate)
				}), mock.MatchedBy(func(t *time.Time) bool {
					return t != nil && t.Equal(endDate)
				})).Return(mockPVZData, nil)
			},
			expectedStatus: http.StatusOK,
			expectedResp:   mockPVZData,
		},
		{
			name:     "Handle invalid date format gracefully",
			userRole: "employee",
			queryParams: map[string]string{
				"limit":     "10",
				"page":      "1",
				"startDate": "invalid-date",
				"endDate":   "invalid-date",
			},
			setupMock: func(mockRepo *mockPVZRepository) {
				mockRepo.On("ListPVZ", 10, 0, (*time.Time)(nil), (*time.Time)(nil)).Return(mockPVZData, nil)
			},
			expectedStatus: http.StatusOK,
			expectedResp:   mockPVZData,
		},
		{
			name:     "Internal server error",
			userRole: "employee",
			queryParams: map[string]string{
				"limit": "10",
				"page":  "1",
			},
			setupMock: func(mockRepo *mockPVZRepository) {
				mockRepo.On("ListPVZ", 10, 0, (*time.Time)(nil), (*time.Time)(nil)).Return(nil, errors.New("database error"))
			},
			expectedStatus: http.StatusInternalServerError,
			expectedResp:   response.ErrorResponse{Message: "Internal server error"},
		},
		{
			name:     "Access denied - invalid role",
			userRole: "guest",
			queryParams: map[string]string{
				"limit": "10",
				"page":  "1",
			},
			expectedStatus: http.StatusForbidden,
			expectedResp:   response.ErrorResponse{Message: "Access denied"},
		},
		{
			name:        "Default pagination when not specified",
			userRole:    "employee",
			queryParams: map[string]string{},
			setupMock: func(mockRepo *mockPVZRepository) {
				mockRepo.On("ListPVZ", 10, 0, (*time.Time)(nil), (*time.Time)(nil)).Return(mockPVZData, nil)
			},
			expectedStatus: http.StatusOK,
			expectedResp:   mockPVZData,
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

			req := httptest.NewRequest(http.MethodGet, "/pvz", nil)
			q := req.URL.Query()
			for key, value := range tt.queryParams {
				q.Add(key, value)
			}
			req.URL.RawQuery = q.Encode()

			ctx := createUserContext(tt.userRole)
			req = req.WithContext(ctx)

			w := httptest.NewRecorder()

			handler.ServeHTTP(w, req)

			require.Equal(t, tt.expectedStatus, w.Code)

			body, err := io.ReadAll(w.Body)
			require.NoError(t, err)

			if tt.expectedStatus == http.StatusOK {
				var pvzResp []*models.PVZ
				err = json.Unmarshal(body, &pvzResp)
				require.NoError(t, err)

				expectedJSON, _ := json.Marshal(tt.expectedResp)
				var expectedPVZs []*models.PVZ
				json.Unmarshal(expectedJSON, &expectedPVZs)

				require.Equal(t, len(expectedPVZs), len(pvzResp))
				for i, expectedPVZ := range expectedPVZs {
					require.Equal(t, expectedPVZ.ID, pvzResp[i].ID)
					require.Equal(t, expectedPVZ.City, pvzResp[i].City)

					require.WithinDuration(t, expectedPVZ.RegistrationDate, pvzResp[i].RegistrationDate, time.Second)
				}
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
