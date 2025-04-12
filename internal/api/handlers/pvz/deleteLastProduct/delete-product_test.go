package deleteLastProduct

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

type mockProductRepository struct {
	mock.Mock
}

func (m *mockProductRepository) AddProduct(product *models.Product) error {
	args := m.Called(product)
	return args.Error(0)
}

func (m *mockProductRepository) GetLastProduct(receptionID string) (*models.Product, error) {
	args := m.Called(receptionID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*models.Product), args.Error(1)
}

func (m *mockProductRepository) DeleteProduct(id string) error {
	args := m.Called(id)
	return args.Error(0)
}

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

func TestDeleteLastProductHandler(t *testing.T) {
	tests := []struct {
		name           string
		pvzID          string
		userRole       string
		setupMock      func(mockProductRepo *mockProductRepository, mockReceptionRepo *mockReceptionRepository)
		expectedStatus int
		expectedResp   interface{}
	}{
		{
			name:     "Successful delete",
			pvzID:    "test-pvz",
			userRole: "employee",
			setupMock: func(mockProductRepo *mockProductRepository, mockReceptionRepo *mockReceptionRepository) {
				mockReceptionRepo.On("GetActiveReception", "test-pvz").Return(&models.Reception{
					ID:    "reception-id",
					PvzID: "test-pvz",
				}, nil)

				product := &models.Product{
					ID:          "product-id",
					DateTime:    time.Now(),
					Type:        "электроника",
					ReceptionID: "reception-id",
				}

				mockProductRepo.On("GetLastProduct", "reception-id").Return(product, nil)
				mockProductRepo.On("DeleteProduct", "product-id").Return(nil)
			},
			expectedStatus: http.StatusOK,
			expectedResp:   nil,
		},
		{
			name:     "No active reception",
			pvzID:    "test-pvz",
			userRole: "employee",
			setupMock: func(mockProductRepo *mockProductRepository, mockReceptionRepo *mockReceptionRepository) {
				mockReceptionRepo.On("GetActiveReception", "test-pvz").Return(nil, internalErrors.ErrNoActiveReception)
			},
			expectedStatus: http.StatusBadRequest,
			expectedResp:   response.ErrorResponse{Message: "No active reception"},
		},
		{
			name:     "No products in reception",
			pvzID:    "test-pvz",
			userRole: "employee",
			setupMock: func(mockProductRepo *mockProductRepository, mockReceptionRepo *mockReceptionRepository) {
				mockReceptionRepo.On("GetActiveReception", "test-pvz").Return(&models.Reception{
					ID:    "reception-id",
					PvzID: "test-pvz",
				}, nil)

				mockProductRepo.On("GetLastProduct", "reception-id").Return(nil, internalErrors.ErrProductNotFound)
			},
			expectedStatus: http.StatusBadRequest,
			expectedResp:   response.ErrorResponse{Message: "No products in reception"},
		},
		{
			name:     "Internal server error",
			pvzID:    "test-pvz",
			userRole: "employee",
			setupMock: func(mockProductRepo *mockProductRepository, mockReceptionRepo *mockReceptionRepository) {
				mockReceptionRepo.On("GetActiveReception", "test-pvz").Return(&models.Reception{
					ID:    "reception-id",
					PvzID: "test-pvz",
				}, nil)

				product := &models.Product{
					ID:          "product-id",
					DateTime:    time.Now(),
					Type:        "электроника",
					ReceptionID: "reception-id",
				}

				mockProductRepo.On("GetLastProduct", "reception-id").Return(product, nil)
				mockProductRepo.On("DeleteProduct", "product-id").Return(errors.New("database error"))
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
			mockProductRepo := new(mockProductRepository)
			mockReceptionRepo := new(mockReceptionRepository)

			if tt.setupMock != nil {
				tt.setupMock(mockProductRepo, mockReceptionRepo)
			}

			productService := services.NewProductService(mockProductRepo, mockReceptionRepo)

			r := chi.NewRouter()
			r.Post("/pvz/{pvzId}/delete_last_product", New(productService))

			req := httptest.NewRequest(http.MethodPost, "/pvz/"+tt.pvzID+"/delete_last_product", nil)

			ctx := createUserContext(tt.userRole)
			req = req.WithContext(ctx)

			w := httptest.NewRecorder()
			r.ServeHTTP(w, req)

			require.Equal(t, tt.expectedStatus, w.Code)

			if tt.expectedResp == nil {
				require.Empty(t, w.Body.String())
			} else {
				var errorResp response.ErrorResponse
				err := json.NewDecoder(w.Body).Decode(&errorResp)
				require.NoError(t, err)
				require.Equal(t, tt.expectedResp, errorResp)
			}

			mockProductRepo.AssertExpectations(t)
			mockReceptionRepo.AssertExpectations(t)
		})
	}
}
