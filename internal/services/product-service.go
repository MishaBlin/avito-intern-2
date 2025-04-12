package services

import (
	"avito-intern/internal/api/dto/internalErrors"
	"avito-intern/internal/api/dto/request/productDto"
	"avito-intern/internal/models"
	"avito-intern/internal/repository"
	"time"

	"github.com/google/uuid"
)

type ProductService struct {
	productRepo   repository.ProductRepositoryInterface
	receptionRepo repository.ReceptionRepositoryInterface
}

func NewProductService(productRepo repository.ProductRepositoryInterface, receptionRepo repository.ReceptionRepositoryInterface) *ProductService {
	return &ProductService{
		productRepo:   productRepo,
		receptionRepo: receptionRepo,
	}
}

func (s *ProductService) AddProduct(req *productDto.CreateProductRequest) (*models.Product, error) {
	reception, err := s.receptionRepo.GetActiveReception(req.PvzID)
	if err != nil {
		return nil, err
	}

	if err = checkProductType(req.Type); err != nil {
		return nil, err
	}

	product := &models.Product{
		ID:          uuid.New().String(),
		DateTime:    time.Now(),
		Type:        req.Type,
		ReceptionID: reception.ID,
	}
	err = s.productRepo.AddProduct(product)
	if err != nil {
		return nil, err
	}
	return product, nil
}

func (s *ProductService) DeleteLastProduct(pvzId string) error {
	reception, err := s.receptionRepo.GetActiveReception(pvzId)
	if err != nil {
		return err
	}
	product, err := s.productRepo.GetLastProduct(reception.ID)
	if err != nil {
		return err
	}
	return s.productRepo.DeleteProduct(product.ID)
}

func checkProductType(productType string) error {
	validTypes := map[string]bool{
		"электроника": true,
		"одежда":      true,
		"обувь":       true,
	}

	if !validTypes[productType] {
		return internalErrors.ErrInvalidProductType
	}
	return nil
}
