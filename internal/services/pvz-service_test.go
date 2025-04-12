package services

import (
	"avito-intern/internal/api/dto/internalErrors"
	"avito-intern/internal/models"
	"errors"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

type mockPVZRepository struct {
	pvzs       map[string]*models.PVZ
	lastLimit  int
	lastOffset int
	createErr  error
	listErr    error
}

func (m *mockPVZRepository) CreatePVZ(pvz *models.PVZ) error {
	if m.createErr != nil {
		return m.createErr
	}
	m.pvzs[pvz.ID] = pvz
	return nil
}

func (m *mockPVZRepository) ListPVZ(limit, offset int, startDate, endDate *time.Time) ([]*models.PVZ, error) {
	if m.listErr != nil {
		return nil, m.listErr
	}
	m.lastLimit = limit
	m.lastOffset = offset
	var result []*models.PVZ
	for _, pvz := range m.pvzs {
		if startDate != nil && pvz.RegistrationDate.Before(*startDate) {
			continue
		}
		if endDate != nil && pvz.RegistrationDate.After(*endDate) {
			continue
		}
		result = append(result, pvz)
	}
	return result, nil
}

func TestPVZService_CreatePVZ_ValidMoscow(t *testing.T) {

	mockRepo := &mockPVZRepository{
		pvzs: make(map[string]*models.PVZ),
	}
	service := NewPVZService(mockRepo)
	pvz := &models.PVZ{
		City:             "Москва",
		RegistrationDate: time.Now(),
	}

	err := service.CreatePVZ(pvz)

	assert.NoError(t, err)
	assert.NotEmpty(t, pvz.ID)
	assert.Equal(t, 1, len(mockRepo.pvzs))
}

func TestPVZService_CreatePVZ_ValidSaintPetersburg(t *testing.T) {
	mockRepo := &mockPVZRepository{
		pvzs: make(map[string]*models.PVZ),
	}
	service := NewPVZService(mockRepo)
	pvz := &models.PVZ{
		City:             "Санкт-Петербург",
		RegistrationDate: time.Now(),
	}

	err := service.CreatePVZ(pvz)

	assert.NoError(t, err)
	assert.NotEmpty(t, pvz.ID)
	assert.Equal(t, 1, len(mockRepo.pvzs))
}

func TestPVZService_CreatePVZ_InvalidCity(t *testing.T) {
	mockRepo := &mockPVZRepository{
		pvzs: make(map[string]*models.PVZ),
	}
	service := NewPVZService(mockRepo)
	pvz := &models.PVZ{
		City:             "Invalid City",
		RegistrationDate: time.Now(),
	}

	err := service.CreatePVZ(pvz)

	assert.Error(t, err)
	assert.Equal(t, internalErrors.ErrInvalidCity, err)
	assert.Equal(t, 0, len(mockRepo.pvzs))
}

func TestPVZService_CreatePVZ_RepositoryError(t *testing.T) {
	mockRepo := &mockPVZRepository{
		pvzs:      make(map[string]*models.PVZ),
		createErr: errors.New("database error"),
	}
	service := NewPVZService(mockRepo)
	pvz := &models.PVZ{
		City:             "Москва",
		RegistrationDate: time.Now(),
	}

	err := service.CreatePVZ(pvz)

	assert.Error(t, err)
	assert.Equal(t, "database error", err.Error())
	assert.Equal(t, 0, len(mockRepo.pvzs))
}

func TestPVZService_ListPVZ_DefaultValues(t *testing.T) {
	now := time.Now()
	mockRepo := &mockPVZRepository{
		pvzs: map[string]*models.PVZ{
			"1": {ID: "1", City: "Москва", RegistrationDate: now},
			"2": {ID: "2", City: "Санкт-Петербург", RegistrationDate: now},
			"3": {ID: "3", City: "Москва", RegistrationDate: now},
		},
	}
	service := NewPVZService(mockRepo)

	pvzs, err := service.ListPVZ("", "", "", "")

	assert.NoError(t, err)
	assert.Equal(t, 3, len(pvzs))
	assert.Equal(t, 10, mockRepo.lastLimit)
	assert.Equal(t, 0, mockRepo.lastOffset)
}

func TestPVZService_ListPVZ_CustomLimitAndPage(t *testing.T) {
	now := time.Now()
	mockRepo := &mockPVZRepository{
		pvzs: map[string]*models.PVZ{
			"1": {ID: "1", City: "Москва", RegistrationDate: now},
			"2": {ID: "2", City: "Санкт-Петербург", RegistrationDate: now},
			"3": {ID: "3", City: "Москва", RegistrationDate: now},
		},
	}
	service := NewPVZService(mockRepo)

	pvzs, err := service.ListPVZ("2", "2", "", "")

	assert.NoError(t, err)
	assert.Equal(t, 3, len(pvzs))
	assert.Equal(t, 2, mockRepo.lastLimit)
	assert.Equal(t, 2, mockRepo.lastOffset)
}

func TestPVZService_ListPVZ_InvalidLimit(t *testing.T) {
	// Setup
	now := time.Now()
	mockRepo := &mockPVZRepository{
		pvzs: map[string]*models.PVZ{
			"1": {ID: "1", City: "Москва", RegistrationDate: now},
			"2": {ID: "2", City: "Санкт-Петербург", RegistrationDate: now},
		},
	}
	service := NewPVZService(mockRepo)

	pvzs, err := service.ListPVZ("invalid", "1", "", "")

	assert.NoError(t, err)
	assert.Equal(t, 2, len(pvzs))
	assert.Equal(t, 10, mockRepo.lastLimit)
	assert.Equal(t, 0, mockRepo.lastOffset)
}

func TestPVZService_ListPVZ_InvalidPage(t *testing.T) {
	now := time.Now()
	mockRepo := &mockPVZRepository{
		pvzs: map[string]*models.PVZ{
			"1": {ID: "1", City: "Москва", RegistrationDate: now},
			"2": {ID: "2", City: "Санкт-Петербург", RegistrationDate: now},
		},
	}
	service := NewPVZService(mockRepo)

	pvzs, err := service.ListPVZ("10", "invalid", "", "")

	assert.NoError(t, err)
	assert.Equal(t, 2, len(pvzs))
	assert.Equal(t, 10, mockRepo.lastLimit)
	assert.Equal(t, 0, mockRepo.lastOffset)
}

func TestPVZService_ListPVZ_DateRange(t *testing.T) {
	now := time.Now()
	mockRepo := &mockPVZRepository{
		pvzs: map[string]*models.PVZ{
			"1": {ID: "1", City: "Москва", RegistrationDate: now},
			"2": {ID: "2", City: "Санкт-Петербург", RegistrationDate: now.Add(-48 * time.Hour)},
			"3": {ID: "3", City: "Москва", RegistrationDate: now.Add(48 * time.Hour)},
		},
	}
	service := NewPVZService(mockRepo)

	startDateStr := now.Add(-24 * time.Hour).Format("2006-01-02T15:04:05")
	endDateStr := now.Add(24 * time.Hour).Format("2006-01-02T15:04:05")

	pvzs, err := service.ListPVZ("10", "1", startDateStr, endDateStr)

	assert.NoError(t, err)
	assert.Equal(t, 1, len(pvzs))
	assert.Equal(t, 10, mockRepo.lastLimit)
	assert.Equal(t, 0, mockRepo.lastOffset)
}

func TestPVZService_ListPVZ_RepositoryError(t *testing.T) {
	mockRepo := &mockPVZRepository{
		pvzs:    make(map[string]*models.PVZ),
		listErr: errors.New("database error"),
	}
	service := NewPVZService(mockRepo)

	pvzs, err := service.ListPVZ("10", "1", "", "")

	assert.Error(t, err)
	assert.Equal(t, "database error", err.Error())
	assert.Nil(t, pvzs)
}
