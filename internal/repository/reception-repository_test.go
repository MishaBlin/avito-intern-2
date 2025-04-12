package repository

import (
	"avito-intern/internal/api/dto/internalErrors"
	"avito-intern/internal/models"
	"database/sql"
	"testing"
	"time"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/stretchr/testify/assert"
)

func TestReceptionRepository_CreateReception_Success(t *testing.T) {

	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("an error '%s' was not expected when opening a stub database connection", err)
	}
	defer db.Close()

	repo := NewReceptionRepository(db)

	reception := &models.Reception{
		ID:       "test-id",
		DateTime: time.Now(),
		PvzID:    "test-pvz",
		Status:   "in_progress",
	}

	mock.ExpectExec("INSERT INTO receptions").
		WithArgs("test-id", sqlmock.AnyArg(), "test-pvz", "in_progress").
		WillReturnResult(sqlmock.NewResult(1, 1))

	err = repo.CreateReception(reception)

	assert.NoError(t, err)
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestReceptionRepository_CreateReception_DatabaseError(t *testing.T) {

	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("an error '%s' was not expected when opening a stub database connection", err)
	}
	defer db.Close()

	repo := NewReceptionRepository(db)

	reception := &models.Reception{
		ID:       "test-id",
		DateTime: time.Now(),
		PvzID:    "test-pvz",
		Status:   "in_progress",
	}

	mock.ExpectExec("INSERT INTO receptions").
		WithArgs("test-id", sqlmock.AnyArg(), "test-pvz", "in_progress").
		WillReturnError(sql.ErrConnDone)

	err = repo.CreateReception(reception)

	assert.Error(t, err)
	assert.Equal(t, sql.ErrConnDone, err)
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestReceptionRepository_GetActiveReception_Success(t *testing.T) {

	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("an error '%s' was not expected when opening a stub database connection", err)
	}
	defer db.Close()

	repo := NewReceptionRepository(db)

	now := time.Now()

	rows := sqlmock.NewRows([]string{"id", "dateTime", "pvzId", "status"}).
		AddRow("test-id", now, "test-pvz", "in_progress")
	mock.ExpectQuery("SELECT id, dateTime, pvzId, status FROM receptions").
		WithArgs("test-pvz", "in_progress").
		WillReturnRows(rows)

	reception, err := repo.GetActiveReception("test-pvz")

	assert.NoError(t, err)
	assert.NotNil(t, reception)
	assert.Equal(t, "test-id", reception.ID)
	assert.Equal(t, "test-pvz", reception.PvzID)
	assert.Equal(t, "in_progress", reception.Status)
	assert.Equal(t, now.Round(time.Second), reception.DateTime.Round(time.Second))
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestReceptionRepository_GetActiveReception_NotFound(t *testing.T) {

	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("an error '%s' was not expected when opening a stub database connection", err)
	}
	defer db.Close()

	repo := NewReceptionRepository(db)

	mock.ExpectQuery("SELECT id, dateTime, pvzId, status FROM receptions").
		WithArgs("test-pvz", "in_progress").
		WillReturnError(sql.ErrNoRows)

	reception, err := repo.GetActiveReception("test-pvz")

	assert.Error(t, err)
	assert.Equal(t, internalErrors.ErrNoActiveReception, err)
	assert.Nil(t, reception)
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestReceptionRepository_GetActiveReception_DatabaseError(t *testing.T) {

	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("an error '%s' was not expected when opening a stub database connection", err)
	}
	defer db.Close()

	repo := NewReceptionRepository(db)

	mock.ExpectQuery("SELECT id, dateTime, pvzId, status FROM receptions").
		WithArgs("test-pvz", "in_progress").
		WillReturnError(sql.ErrConnDone)

	reception, err := repo.GetActiveReception("test-pvz")

	assert.Error(t, err)
	assert.Equal(t, sql.ErrConnDone, err)
	assert.Nil(t, reception)
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestReceptionRepository_CloseReception_Success(t *testing.T) {

	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("an error '%s' was not expected when opening a stub database connection", err)
	}
	defer db.Close()

	repo := NewReceptionRepository(db)

	mock.ExpectExec("UPDATE receptions").
		WithArgs("close", "test-id").
		WillReturnResult(sqlmock.NewResult(1, 1))

	err = repo.CloseReception("test-id")

	assert.NoError(t, err)
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestReceptionRepository_CloseReception_NoRowsAffected(t *testing.T) {

	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("an error '%s' was not expected when opening a stub database connection", err)
	}
	defer db.Close()

	repo := NewReceptionRepository(db)

	mock.ExpectExec("UPDATE receptions").
		WithArgs("close", "nonexistent-id").
		WillReturnResult(sqlmock.NewResult(0, 0))

	err = repo.CloseReception("nonexistent-id")

	assert.Error(t, err)
	assert.Equal(t, "no reception updated", err.Error())
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestReceptionRepository_CloseReception_DatabaseError(t *testing.T) {

	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("an error '%s' was not expected when opening a stub database connection", err)
	}
	defer db.Close()

	repo := NewReceptionRepository(db)

	mock.ExpectExec("UPDATE receptions").
		WithArgs("close", "test-id").
		WillReturnError(sql.ErrConnDone)

	err = repo.CloseReception("test-id")

	assert.Error(t, err)
	assert.Equal(t, sql.ErrConnDone, err)
	assert.NoError(t, mock.ExpectationsWereMet())
}
