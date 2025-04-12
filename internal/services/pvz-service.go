package services

import (
	"avito-intern/internal/api/dto/internalErrors"
	"avito-intern/internal/models"
	"avito-intern/internal/repository"
	"strconv"
	"time"

	"github.com/google/uuid"
)

type PVZService struct {
	pvzRepo repository.PVZRepositoryInterface
}

func NewPVZService(pvzRepo repository.PVZRepositoryInterface) *PVZService {
	return &PVZService{
		pvzRepo: pvzRepo,
	}
}

func (s *PVZService) CreatePVZ(pvz *models.PVZ) error {
	if err := checkCity(pvz.City); err != nil {
		return err
	}

	if pvz.ID == "" {
		pvz.ID = uuid.New().String()
	}
	if pvz.RegistrationDate.IsZero() {
		pvz.RegistrationDate = time.Now()
	}
	return s.pvzRepo.CreatePVZ(pvz)
}

func (s *PVZService) ListPVZ(limitStr, pageStr, startDateStr, endDateStr string) ([]*models.PVZ, error) {
	page, _ := strconv.Atoi(pageStr)
	if page < 1 {
		page = 1
	}
	limit, _ := strconv.Atoi(limitStr)
	if limit < 1 || limit > 30 {
		limit = 10
	}
	offset := (page - 1) * limit

	var startDate, endDate *time.Time
	layout := "2006-01-02T15:04:05"
	if startDateStr != "" {
		if t, err := time.Parse(layout, startDateStr); err == nil {
			startDate = &t
		}
	}
	if endDateStr != "" {
		if t, err := time.Parse(layout, endDateStr); err == nil {
			endDate = &t
		}
	}

	return s.pvzRepo.ListPVZ(limit, offset, startDate, endDate)
}

func checkCity(city string) error {
	allowedCities := map[string]bool{
		"Москва":          true,
		"Санкт-Петербург": true,
		"Казань":          true,
	}

	if !allowedCities[city] {
		return internalErrors.ErrInvalidCity
	}
	return nil
}
