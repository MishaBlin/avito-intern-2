package services

import (
	"avito-intern/internal/api/dto/internalErrors"
	"avito-intern/internal/models"
	"errors"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

type mockReceptionServiceRepository struct {
	receptions map[string]*models.Reception
	createErr  error
	getErr     error
	closeErr   error
}

func (m *mockReceptionServiceRepository) CreateReception(reception *models.Reception) error {
	if m.createErr != nil {
		return m.createErr
	}
	if _, exists := m.receptions[reception.ID]; exists {
		return internalErrors.ErrActiveReceptionExists
	}
	m.receptions[reception.ID] = reception
	return nil
}

func (m *mockReceptionServiceRepository) GetActiveReception(pvzID string) (*models.Reception, error) {
	if m.getErr != nil {
		return nil, m.getErr
	}
	for _, reception := range m.receptions {
		if reception.PvzID == pvzID && reception.Status == "in_progress" {
			return reception, nil
		}
	}
	return nil, internalErrors.ErrNoActiveReception
}

func (m *mockReceptionServiceRepository) CloseReception(receptionID string) error {
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

func TestReceptionService_CreateReception_Success(t *testing.T) {
	mockRepo := &mockReceptionServiceRepository{
		receptions: make(map[string]*models.Reception),
	}
	service := NewReceptionService(mockRepo)
	reception := &models.Reception{
		PvzID:    "test-pvz",
		DateTime: time.Now(),
		Status:   "in_progress",
	}

	err := service.CreateReception(reception)

	assert.NoError(t, err)
	assert.NotEmpty(t, reception.ID)
	assert.Equal(t, 1, len(mockRepo.receptions))
}

func TestReceptionService_CreateReception_ActiveReceptionExists(t *testing.T) {
	existingReception := &models.Reception{
		ID:       "existing-id",
		PvzID:    "test-pvz",
		DateTime: time.Now(),
		Status:   "in_progress",
	}

	mockRepo := &mockReceptionServiceRepository{
		receptions: map[string]*models.Reception{
			"existing-id": existingReception,
		},
	}
	service := NewReceptionService(mockRepo)

	newReception := &models.Reception{
		ID:       "existing-id",
		PvzID:    "test-pvz",
		DateTime: time.Now(),
		Status:   "in_progress",
	}

	err := service.CreateReception(newReception)

	assert.Error(t, err)
	assert.Equal(t, internalErrors.ErrActiveReceptionExists, err)
}

func TestReceptionService_CreateReception_RepositoryError(t *testing.T) {
	mockRepo := &mockReceptionServiceRepository{
		receptions: make(map[string]*models.Reception),
		createErr:  errors.New("database error"),
	}
	service := NewReceptionService(mockRepo)
	reception := &models.Reception{
		PvzID:    "test-pvz",
		DateTime: time.Now(),
		Status:   "in_progress",
	}

	err := service.CreateReception(reception)

	assert.Error(t, err)
	assert.Equal(t, "database error", err.Error())
}

func TestReceptionService_CloseLastReception_Success(t *testing.T) {
	now := time.Now()
	activeReception := &models.Reception{
		ID:       "active-reception",
		PvzID:    "test-pvz",
		DateTime: now,
		Status:   "in_progress",
	}

	mockRepo := &mockReceptionServiceRepository{
		receptions: map[string]*models.Reception{
			"active-reception": activeReception,
		},
	}
	service := NewReceptionService(mockRepo)

	reception, err := service.CloseLastReception("test-pvz")

	assert.NoError(t, err)
	assert.NotNil(t, reception)
	assert.Equal(t, "close", reception.Status)
	assert.Equal(t, "active-reception", reception.ID)
}

func TestReceptionService_CloseLastReception_NoActiveReception(t *testing.T) {
	mockRepo := &mockReceptionServiceRepository{
		receptions: make(map[string]*models.Reception),
	}
	service := NewReceptionService(mockRepo)

	reception, err := service.CloseLastReception("test-pvz")

	assert.Error(t, err)
	assert.Equal(t, internalErrors.ErrNoActiveReception, err)
	assert.Nil(t, reception)
}

func TestReceptionService_CloseLastReception_GetActiveReceptionError(t *testing.T) {
	mockRepo := &mockReceptionServiceRepository{
		receptions: make(map[string]*models.Reception),
		getErr:     errors.New("database error"),
	}
	service := NewReceptionService(mockRepo)

	reception, err := service.CloseLastReception("test-pvz")

	assert.Error(t, err)
	assert.Equal(t, internalErrors.ErrNoActiveReception, err)
	assert.Nil(t, reception)
}

func TestReceptionService_CloseLastReception_CloseReceptionError(t *testing.T) {
	now := time.Now()
	activeReception := &models.Reception{
		ID:       "active-reception",
		PvzID:    "test-pvz",
		DateTime: now,
		Status:   "in_progress",
	}

	mockRepo := &mockReceptionServiceRepository{
		receptions: map[string]*models.Reception{
			"active-reception": activeReception,
		},
		closeErr: errors.New("database error"),
	}
	service := NewReceptionService(mockRepo)

	reception, err := service.CloseLastReception("test-pvz")

	assert.Error(t, err)
	assert.Equal(t, "database error", err.Error())
	assert.Nil(t, reception)
}
