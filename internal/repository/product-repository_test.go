package repository

import (
	"avito-intern/internal/api/dto/internalErrors"
	"avito-intern/internal/models"
	"database/sql"
	"errors"
	"testing"
	"time"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/stretchr/testify/assert"
)

func TestProductRepository_AddProduct_Success(t *testing.T) {
	// Setup
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("an error '%s' was not expected when opening a stub database connection", err)
	}
	defer db.Close()

	repo := NewProductRepository(db)

	product := &models.Product{
		ID:          "test-id",
		DateTime:    time.Now(),
		Type:        "электроника",
		ReceptionID: "test-reception",
	}

	mock.ExpectExec("INSERT INTO products").
		WithArgs("test-id", sqlmock.AnyArg(), "электроника", "test-reception").
		WillReturnResult(sqlmock.NewResult(1, 1))

	err = repo.AddProduct(product)

	assert.NoError(t, err)
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestProductRepository_AddProduct_DatabaseError(t *testing.T) {

	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("an error '%s' was not expected when opening a stub database connection", err)
	}
	defer db.Close()

	repo := NewProductRepository(db)

	product := &models.Product{
		ID:          "test-id",
		DateTime:    time.Now(),
		Type:        "электроника",
		ReceptionID: "test-reception",
	}

	mock.ExpectExec("INSERT INTO products").
		WithArgs("test-id", sqlmock.AnyArg(), "электроника", "test-reception").
		WillReturnError(sql.ErrConnDone)

	err = repo.AddProduct(product)

	assert.Error(t, err)
	assert.Equal(t, sql.ErrConnDone, err)
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestProductRepository_GetLastProduct_Success(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("an error '%s' was not expected when opening a stub database connection", err)
	}
	defer db.Close()

	repo := NewProductRepository(db)
	now := time.Now()

	rows := sqlmock.NewRows([]string{"id", "dateTime", "type", "receptionId"}).
		AddRow("test-id", now, "электроника", "test-reception")
	mock.ExpectQuery("SELECT id, dateTime, type, receptionId FROM products").
		WithArgs("test-reception").
		WillReturnRows(rows)

	product, err := repo.GetLastProduct("test-reception")

	assert.NoError(t, err)
	assert.NotNil(t, product)
	assert.Equal(t, "test-id", product.ID)
	assert.Equal(t, "электроника", product.Type)
	assert.Equal(t, "test-reception", product.ReceptionID)
	assert.Equal(t, now.Round(time.Second), product.DateTime.Round(time.Second))
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestProductRepository_GetLastProduct_NotFound(t *testing.T) {

	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("an error '%s' was not expected when opening a stub database connection", err)
	}
	defer db.Close()

	repo := NewProductRepository(db)

	mock.ExpectQuery("SELECT id, dateTime, type, receptionId FROM products").
		WithArgs("test-reception").
		WillReturnError(sql.ErrNoRows)

	product, err := repo.GetLastProduct("test-reception")

	assert.Error(t, err)
	assert.Equal(t, internalErrors.ErrProductNotFound, err)
	assert.Nil(t, product)
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestProductRepository_GetLastProduct_DatabaseError(t *testing.T) {

	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("an error '%s' was not expected when opening a stub database connection", err)
	}
	defer db.Close()

	repo := NewProductRepository(db)

	mock.ExpectQuery("SELECT id, dateTime, type, receptionId FROM products").
		WithArgs("test-reception").
		WillReturnError(sql.ErrConnDone)

	product, err := repo.GetLastProduct("test-reception")

	assert.Error(t, err)
	assert.Equal(t, sql.ErrConnDone, err)
	assert.Nil(t, product)
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestProductRepository_DeleteProduct_Success(t *testing.T) {

	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("an error '%s' was not expected when opening a stub database connection", err)
	}
	defer db.Close()

	repo := NewProductRepository(db)

	mock.ExpectExec("DELETE FROM products").
		WithArgs("test-id").
		WillReturnResult(sqlmock.NewResult(1, 1))

	err = repo.DeleteProduct("test-id")

	assert.NoError(t, err)
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestProductRepository_DeleteProduct_NoRowsAffected(t *testing.T) {

	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("an error '%s' was not expected when opening a stub database connection", err)
	}
	defer db.Close()

	repo := NewProductRepository(db)

	mock.ExpectExec("DELETE FROM products").
		WithArgs("nonexistent-id").
		WillReturnResult(sqlmock.NewResult(0, 0))

	err = repo.DeleteProduct("nonexistent-id")

	assert.Error(t, err)
	assert.Equal(t, "no product deleted", err.Error())
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestProductRepository_DeleteProduct_DatabaseError(t *testing.T) {

	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("an error '%s' was not expected when opening a stub database connection", err)
	}
	defer db.Close()

	repo := NewProductRepository(db)

	mock.ExpectExec("DELETE FROM products").
		WithArgs("test-id").
		WillReturnError(sql.ErrConnDone)

	err = repo.DeleteProduct("test-id")

	assert.Error(t, err)
	assert.Equal(t, sql.ErrConnDone, err)
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestProductRepository_DeleteProduct_RowsAffectedError(t *testing.T) {

	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("an error '%s' was not expected when opening a stub database connection", err)
	}
	defer db.Close()

	repo := NewProductRepository(db)

	result := sqlmock.NewErrorResult(errors.New("rows affected error"))

	mock.ExpectExec("DELETE FROM products").
		WithArgs("test-id").
		WillReturnResult(result)

	err = repo.DeleteProduct("test-id")

	assert.Error(t, err)
	assert.Equal(t, "rows affected error", err.Error())
	assert.NoError(t, mock.ExpectationsWereMet())
}
