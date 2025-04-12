package services

import (
	"avito-intern/internal/api/dto/internalErrors"
	"avito-intern/internal/api/dto/request/productDto"
	"avito-intern/internal/models"
	"errors"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

type mockProductRepository struct {
	products  map[string]*models.Product
	addErr    error
	getErr    error
	deleteErr error
}

func (m *mockProductRepository) AddProduct(product *models.Product) error {
	if m.addErr != nil {
		return m.addErr
	}
	m.products[product.ID] = product
	return nil
}

func (m *mockProductRepository) GetLastProduct(receptionID string) (*models.Product, error) {
	if m.getErr != nil {
		return nil, m.getErr
	}
	for _, product := range m.products {
		if product.ReceptionID == receptionID {
			return product, nil
		}
	}
	return nil, internalErrors.ErrProductNotFound
}

func (m *mockProductRepository) DeleteProduct(id string) error {
	if m.deleteErr != nil {
		return m.deleteErr
	}
	if _, exists := m.products[id]; !exists {
		return internalErrors.ErrProductNotFound
	}
	delete(m.products, id)
	return nil
}

type mockReceptionRepository struct {
	receptions map[string]*models.Reception
	getErr     error
	createErr  error
	closeErr   error
}

func (m *mockReceptionRepository) CreateReception(reception *models.Reception) error {
	if m.createErr != nil {
		return m.createErr
	}
	if _, exists := m.receptions[reception.ID]; exists {
		return internalErrors.ErrActiveReceptionExists
	}
	m.receptions[reception.ID] = reception
	return nil
}

func (m *mockReceptionRepository) GetActiveReception(pvzID string) (*models.Reception, error) {
	if m.getErr != nil {
		return nil, m.getErr
	}
	for _, reception := range m.receptions {
		if reception.PvzID == pvzID {
			return reception, nil
		}
	}
	return nil, internalErrors.ErrNoActiveReception
}

func (m *mockReceptionRepository) CloseReception(receptionID string) error {
	if m.closeErr != nil {
		return m.closeErr
	}
	reception, exists := m.receptions[receptionID]
	if !exists {
		return internalErrors.ErrNoActiveReception
	}
	reception.Status = "close"
	return nil
}

func TestProductService_AddProduct_ValidProduct(t *testing.T) {
	mockProductRepo := &mockProductRepository{
		products: make(map[string]*models.Product),
	}
	mockReceptionRepo := &mockReceptionRepository{
		receptions: map[string]*models.Reception{
			"test-reception": {
				ID:    "test-reception",
				PvzID: "test-pvz",
			},
		},
	}

	service := NewProductService(mockProductRepo, mockReceptionRepo)
	req := &productDto.CreateProductRequest{
		PvzID: "test-pvz",
		Type:  "электроника",
	}

	product, err := service.AddProduct(req)

	assert.NoError(t, err)
	assert.NotNil(t, product)
	assert.Equal(t, "электроника", product.Type)
	assert.Equal(t, "test-reception", product.ReceptionID)
}

func TestProductService_AddProduct_InvalidProductType(t *testing.T) {
	mockProductRepo := &mockProductRepository{
		products: make(map[string]*models.Product),
	}
	mockReceptionRepo := &mockReceptionRepository{
		receptions: map[string]*models.Reception{
			"test-reception": {
				ID:    "test-reception",
				PvzID: "test-pvz",
			},
		},
	}

	service := NewProductService(mockProductRepo, mockReceptionRepo)
	req := &productDto.CreateProductRequest{
		PvzID: "test-pvz",
		Type:  "invalid-type",
	}

	product, err := service.AddProduct(req)

	assert.Error(t, err)
	assert.Equal(t, internalErrors.ErrInvalidProductType, err)
	assert.Nil(t, product)
}

func TestProductService_AddProduct_NonExistentPVZ(t *testing.T) {
	mockProductRepo := &mockProductRepository{
		products: make(map[string]*models.Product),
	}
	mockReceptionRepo := &mockReceptionRepository{
		receptions: map[string]*models.Reception{
			"test-reception": {
				ID:    "test-reception",
				PvzID: "test-pvz",
			},
		},
	}

	service := NewProductService(mockProductRepo, mockReceptionRepo)
	req := &productDto.CreateProductRequest{
		PvzID: "non-existent-pvz",
		Type:  "электроника",
	}

	product, err := service.AddProduct(req)

	assert.Error(t, err)
	assert.Equal(t, internalErrors.ErrNoActiveReception, err)
	assert.Nil(t, product)
}

func TestProductService_AddProduct_RepositoryError(t *testing.T) {

	mockProductRepo := &mockProductRepository{
		products: make(map[string]*models.Product),
		addErr:   errors.New("database error"),
	}
	mockReceptionRepo := &mockReceptionRepository{
		receptions: map[string]*models.Reception{
			"test-reception": {
				ID:    "test-reception",
				PvzID: "test-pvz",
			},
		},
	}

	service := NewProductService(mockProductRepo, mockReceptionRepo)
	req := &productDto.CreateProductRequest{
		PvzID: "test-pvz",
		Type:  "электроника",
	}

	product, err := service.AddProduct(req)

	assert.Error(t, err)
	assert.Equal(t, "database error", err.Error())
	assert.Nil(t, product)
}

func TestProductService_DeleteLastProduct_Success(t *testing.T) {

	mockProductRepo := &mockProductRepository{
		products: map[string]*models.Product{
			"test-product": {
				ID:          "test-product",
				DateTime:    time.Now(),
				Type:        "электроника",
				ReceptionID: "test-reception",
			},
		},
	}
	mockReceptionRepo := &mockReceptionRepository{
		receptions: map[string]*models.Reception{
			"test-reception": {
				ID:    "test-reception",
				PvzID: "test-pvz",
			},
		},
	}

	service := NewProductService(mockProductRepo, mockReceptionRepo)

	err := service.DeleteLastProduct("test-pvz")

	assert.NoError(t, err)
	assert.Empty(t, mockProductRepo.products)
}

func TestProductService_DeleteLastProduct_NonExistentPVZ(t *testing.T) {

	mockProductRepo := &mockProductRepository{
		products: map[string]*models.Product{
			"test-product": {
				ID:          "test-product",
				DateTime:    time.Now(),
				Type:        "электроника",
				ReceptionID: "test-reception",
			},
		},
	}
	mockReceptionRepo := &mockReceptionRepository{
		receptions: map[string]*models.Reception{
			"test-reception": {
				ID:    "test-reception",
				PvzID: "test-pvz",
			},
		},
	}

	service := NewProductService(mockProductRepo, mockReceptionRepo)

	err := service.DeleteLastProduct("non-existent-pvz")

	assert.Error(t, err)
	assert.Equal(t, internalErrors.ErrNoActiveReception, err)
	assert.Len(t, mockProductRepo.products, 1)
}

func TestProductService_DeleteLastProduct_NoProductsInReception(t *testing.T) {

	mockProductRepo := &mockProductRepository{
		products: make(map[string]*models.Product),
	}
	mockReceptionRepo := &mockReceptionRepository{
		receptions: map[string]*models.Reception{
			"test-reception": {
				ID:    "test-reception",
				PvzID: "test-pvz",
			},
		},
	}

	service := NewProductService(mockProductRepo, mockReceptionRepo)

	err := service.DeleteLastProduct("test-pvz")

	assert.Error(t, err)
	assert.Equal(t, internalErrors.ErrProductNotFound, err)
}

func TestProductService_DeleteLastProduct_DeleteError(t *testing.T) {

	mockProductRepo := &mockProductRepository{
		products: map[string]*models.Product{
			"test-product": {
				ID:          "test-product",
				DateTime:    time.Now(),
				Type:        "электроника",
				ReceptionID: "test-reception",
			},
		},
		deleteErr: errors.New("delete error"),
	}
	mockReceptionRepo := &mockReceptionRepository{
		receptions: map[string]*models.Reception{
			"test-reception": {
				ID:    "test-reception",
				PvzID: "test-pvz",
			},
		},
	}

	service := NewProductService(mockProductRepo, mockReceptionRepo)

	err := service.DeleteLastProduct("test-pvz")

	assert.Error(t, err)
	assert.Equal(t, "delete error", err.Error())
	assert.Len(t, mockProductRepo.products, 1)
}
