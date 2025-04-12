package repository

import (
	"avito-intern/internal/models"
	"database/sql"
	"time"

	"github.com/Masterminds/squirrel"
)

type PVZRepositoryInterface interface {
	CreatePVZ(pvz *models.PVZ) error
	ListPVZ(limit, offset int, startDate, endDate *time.Time) ([]*models.PVZ, error)
}

type PVZRepository struct {
	db         *sql.DB
	sqlBuilder squirrel.StatementBuilderType
}

func NewPVZRepository(db *sql.DB) *PVZRepository {
	return &PVZRepository{
		db:         db,
		sqlBuilder: squirrel.StatementBuilder.PlaceholderFormat(squirrel.Dollar),
	}
}

func (r *PVZRepository) CreatePVZ(pvz *models.PVZ) error {
	query, args, err := r.sqlBuilder.
		Insert("pvz").
		Columns("id", "registrationDate", "city").
		Values(pvz.ID, pvz.RegistrationDate, pvz.City).
		ToSql()
	if err != nil {
		return err
	}
	_, err = r.db.Exec(query, args...)
	return err
}

func (r *PVZRepository) ListPVZ(limit, offset int, startDate, endDate *time.Time) ([]*models.PVZ, error) {
	q := r.sqlBuilder.
		Select("id", "registrationDate", "city").
		From("pvz")
	if startDate != nil {
		q = q.Where("registrationDate >= ?", *startDate)
	}
	if endDate != nil {
		q = q.Where("registrationDate <= ?", *endDate)
	}
	q = q.Limit(uint64(limit)).Offset(uint64(offset))

	query, args, err := q.ToSql()
	if err != nil {
		return nil, err
	}
	rows, err := r.db.Query(query, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	pvzs := make([]*models.PVZ, 0)
	for rows.Next() {
		var pvz models.PVZ
		if err := rows.Scan(&pvz.ID, &pvz.RegistrationDate, &pvz.City); err != nil {
			continue
		}
		pvzs = append(pvzs, &pvz)
	}
	return pvzs, nil
}
