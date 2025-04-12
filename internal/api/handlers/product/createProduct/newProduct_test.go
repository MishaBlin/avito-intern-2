package createProduct

import (
	"avito-intern/internal/api/dto/internalErrors"
	"avito-intern/internal/api/dto/request/productDto"
	"avito-intern/internal/api/dto/response"
	"avito-intern/internal/api/middleware"
	"avito-intern/internal/models"
	"avito-intern/internal/services"
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

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

func TestCreateProductHandler(t *testing.T) {
	tests := []struct {
		name           string
		productReq     productDto.CreateProductRequest
		userRole       string
		invalidBody    bool
		setupMock      func(mockProductRepo *mockProductRepository, mockReceptionRepo *mockReceptionRepository)
		expectedStatus int
		expectedResp   interface{}
	}{
		{
			name: "Successful product creation",
			productReq: productDto.CreateProductRequest{
				Type:  "электроника",
				PvzID: "test-pvz",
			},
			userRole: "employee",
			setupMock: func(mockProductRepo *mockProductRepository, mockReceptionRepo *mockReceptionRepository) {
				mockReceptionRepo.On("GetActiveReception", "test-pvz").Return(&models.Reception{
					ID:    "reception-id",
					PvzID: "test-pvz",
				}, nil)

				mockProductRepo.On("AddProduct", mock.MatchedBy(func(product *models.Product) bool {
					return product.Type == "электроника" && product.ReceptionID == "reception-id"
				})).Return(nil)
			},
			expectedStatus: http.StatusCreated,
		},
		{
			name: "No active reception",
			productReq: productDto.CreateProductRequest{
				Type:  "электроника",
				PvzID: "test-pvz",
			},
			userRole: "employee",
			setupMock: func(mockProductRepo *mockProductRepository, mockReceptionRepo *mockReceptionRepository) {
				mockReceptionRepo.On("GetActiveReception", "test-pvz").Return(nil, internalErrors.ErrNoActiveReception)
			},
			expectedStatus: http.StatusBadRequest,
			expectedResp:   response.ErrorResponse{Message: "No active reception"},
		},
		{
			name: "Invalid product type",
			productReq: productDto.CreateProductRequest{
				Type:  "invalid-type",
				PvzID: "test-pvz",
			},
			userRole: "employee",
			setupMock: func(mockProductRepo *mockProductRepository, mockReceptionRepo *mockReceptionRepository) {
				mockReceptionRepo.On("GetActiveReception", "test-pvz").Return(&models.Reception{
					ID:    "reception-id",
					PvzID: "test-pvz",
				}, nil)
			},
			expectedStatus: http.StatusBadRequest,
			expectedResp:   response.ErrorResponse{Message: "Invalid product type"},
		},
		{
			name:           "Invalid request body",
			invalidBody:    true,
			userRole:       "employee",
			expectedStatus: http.StatusBadRequest,
			expectedResp:   response.ErrorResponse{Message: "Invalid request"},
		},
		{
			name: "Empty PvzID",
			productReq: productDto.CreateProductRequest{
				Type:  "электроника",
				PvzID: "",
			},
			userRole:       "employee",
			expectedStatus: http.StatusBadRequest,
			expectedResp:   response.ErrorResponse{Message: "Invalid request"},
		},
		{
			name: "Access denied",
			productReq: productDto.CreateProductRequest{
				Type:  "электроника",
				PvzID: "test-pvz",
			},
			userRole:       "user",
			expectedStatus: http.StatusForbidden,
			expectedResp:   response.ErrorResponse{Message: "Access denied"},
		},
		{
			name: "Server error",
			productReq: productDto.CreateProductRequest{
				Type:  "электроника",
				PvzID: "test-pvz",
			},
			userRole: "employee",
			setupMock: func(mockProductRepo *mockProductRepository, mockReceptionRepo *mockReceptionRepository) {
				mockReceptionRepo.On("GetActiveReception", "test-pvz").Return(&models.Reception{
					ID:    "reception-id",
					PvzID: "test-pvz",
				}, nil)

				mockProductRepo.On("AddProduct", mock.MatchedBy(func(product *models.Product) bool {
					return product.Type == "электроника" && product.ReceptionID == "reception-id"
				})).Return(errors.New("database error"))
			},
			expectedStatus: http.StatusInternalServerError,
			expectedResp:   response.ErrorResponse{Message: "Invalid product type"},
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
			handler := New(productService)

			var req *http.Request
			if tt.invalidBody {
				req = httptest.NewRequest(http.MethodPost, "/products", strings.NewReader("invalid json"))
			} else {
				body, _ := json.Marshal(tt.productReq)
				req = httptest.NewRequest(http.MethodPost, "/products", bytes.NewReader(body))
			}

			ctx := createUserContext(tt.userRole)
			req = req.WithContext(ctx)

			req.Header.Set("Content-Type", "application/json")
			w := httptest.NewRecorder()

			handler.ServeHTTP(w, req)

			require.Equal(t, tt.expectedStatus, w.Code)

			if tt.expectedStatus == http.StatusCreated {
				var product models.Product
				err := json.NewDecoder(w.Body).Decode(&product)
				require.NoError(t, err)

				require.Equal(t, tt.productReq.Type, product.Type)
				require.NotEmpty(t, product.ID)
				require.NotEmpty(t, product.ReceptionID)
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
