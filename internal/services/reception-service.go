package services

import (
	"avito-intern/internal/api/dto/internalErrors"
	"avito-intern/internal/models"
	"avito-intern/internal/repository"
	"time"

	"github.com/google/uuid"
)

type ReceptionService struct {
	receptionRepo repository.ReceptionRepositoryInterface
}

func NewReceptionService(receptionRepo repository.ReceptionRepositoryInterface) *ReceptionService {
	return &ReceptionService{
		receptionRepo: receptionRepo,
	}
}

func (s *ReceptionService) CreateReception(reception *models.Reception) error {
	activeReception, _ := s.getActiveReception(reception.PvzID)
	if activeReception != nil {
		return internalErrors.ErrActiveReceptionExists
	}
	if reception.ID == "" {
		reception.ID = uuid.New().String()
	}
	if reception.DateTime.IsZero() {
		reception.DateTime = time.Now()
	}
	return s.receptionRepo.CreateReception(reception)
}

func (s *ReceptionService) getActiveReception(pvzId string) (*models.Reception, error) {
	return s.receptionRepo.GetActiveReception(pvzId)
}

func (s *ReceptionService) CloseLastReception(pvzID string) (*models.Reception, error) {
	reception, err := s.getActiveReception(pvzID)
	if err != nil {
		return nil, internalErrors.ErrNoActiveReception
	}
	err = s.receptionRepo.CloseReception(reception.ID)
	if err != nil {
		return nil, err
	}
	reception.Status = "close"
	return reception, nil
}
