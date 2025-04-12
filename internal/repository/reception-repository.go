package repository

import (
	"avito-intern/internal/api/dto/internalErrors"
	"avito-intern/internal/models"
	"database/sql"
	"errors"

	"github.com/Masterminds/squirrel"
)

type ReceptionRepositoryInterface interface {
	CreateReception(reception *models.Reception) error
	GetActiveReception(pvzID string) (*models.Reception, error)
	CloseReception(receptionID string) error
}

type ReceptionRepository struct {
	db         *sql.DB
	sqlBuilder squirrel.StatementBuilderType
}

func NewReceptionRepository(db *sql.DB) *ReceptionRepository {
	return &ReceptionRepository{
		db:         db,
		sqlBuilder: squirrel.StatementBuilder.PlaceholderFormat(squirrel.Dollar),
	}
}

func (r *ReceptionRepository) CreateReception(reception *models.Reception) error {
	query, args, err := r.sqlBuilder.
		Insert("receptions").
		Columns("id", "dateTime", "pvzId", "status").
		Values(reception.ID, reception.DateTime, reception.PvzID, reception.Status).
		ToSql()
	if err != nil {
		return err
	}
	_, err = r.db.Exec(query, args...)
	return err
}

func (r *ReceptionRepository) GetActiveReception(pvzId string) (*models.Reception, error) {
	var reception models.Reception
	query, args, err := r.sqlBuilder.
		Select("id", "dateTime", "pvzId", "status").
		From("receptions").
		Where(squirrel.Eq{"pvzId": pvzId, "status": "in_progress"}).
		OrderBy("dateTime DESC").
		Limit(1).
		ToSql()
	if err != nil {
		return nil, err
	}
	err = r.db.QueryRow(query, args...).Scan(&reception.ID, &reception.DateTime, &reception.PvzID, &reception.Status)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, internalErrors.ErrNoActiveReception
		}
		return nil, err
	}
	return &reception, nil
}

func (r *ReceptionRepository) CloseReception(receptionID string) error {
	query, args, err := r.sqlBuilder.
		Update("receptions").
		Set("status", "close").
		Where(squirrel.Eq{"id": receptionID}).
		ToSql()
	if err != nil {
		return err
	}
	res, err := r.db.Exec(query, args...)
	if err != nil {
		return err
	}
	rowsAffected, err := res.RowsAffected()
	if err != nil {
		return err
	}
	if rowsAffected == 0 {
		return errors.New("no reception updated")
	}
	return nil
}
