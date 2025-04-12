package repository

import (
	"avito-intern/internal/api/dto/internalErrors"
	"avito-intern/internal/models"
	"database/sql"
	"errors"
	"time"

	"github.com/Masterminds/squirrel"
	"github.com/google/uuid"
)

type ProductRepositoryInterface interface {
	AddProduct(product *models.Product) error
	GetLastProduct(receptionID string) (*models.Product, error)
	DeleteProduct(id string) error
}

type ProductRepository struct {
	db         *sql.DB
	sqlBuilder squirrel.StatementBuilderType
}

func NewProductRepository(db *sql.DB) *ProductRepository {
	return &ProductRepository{
		db:         db,
		sqlBuilder: squirrel.StatementBuilder.PlaceholderFormat(squirrel.Dollar),
	}
}

func (r *ProductRepository) AddProduct(product *models.Product) error {
	if product.ID == "" {
		product.ID = uuid.New().String()
	}
	if product.DateTime.IsZero() {
		product.DateTime = time.Now()
	}
	query, args, err := r.sqlBuilder.
		Insert("products").
		Columns("id", "dateTime", "type", "receptionId").
		Values(product.ID, product.DateTime, product.Type, product.ReceptionID).
		ToSql()
	if err != nil {
		return err
	}
	_, err = r.db.Exec(query, args...)
	return err
}

func (r *ProductRepository) GetLastProduct(receptionID string) (*models.Product, error) {
	var product models.Product
	query, args, err := r.sqlBuilder.
		Select("id", "dateTime", "type", "receptionId").
		From("products").
		Where(squirrel.Eq{"receptionId": receptionID}).
		OrderBy("dateTime DESC").
		Limit(1).
		ToSql()
	if err != nil {
		return nil, err
	}
	err = r.db.QueryRow(query, args...).Scan(&product.ID, &product.DateTime, &product.Type, &product.ReceptionID)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, internalErrors.ErrProductNotFound
		}
		return nil, err
	}
	return &product, nil
}

func (r *ProductRepository) DeleteProduct(productID string) error {
	query, args, err := r.sqlBuilder.
		Delete("products").
		Where(squirrel.Eq{"id": productID}).
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
		return errors.New("no product deleted")
	}
	return nil
}
