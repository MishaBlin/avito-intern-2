package repository

import (
	"avito-intern/internal/models"
	"database/sql"
	"testing"
	"time"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/stretchr/testify/assert"
)

func TestPVZRepository_CreatePVZ(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("an error '%s' was not expected when opening a stub database connection", err)
	}
	defer db.Close()

	repo := NewPVZRepository(db)

	tests := []struct {
		name    string
		pvz     *models.PVZ
		mock    func()
		wantErr bool
	}{
		{
			name: "Successfully create PVZ",
			pvz: &models.PVZ{
				ID:               "test-id",
				RegistrationDate: time.Now(),
				City:             "Москва",
			},
			mock: func() {
				mock.ExpectExec("INSERT INTO pvz").
					WithArgs("test-id", sqlmock.AnyArg(), "Москва").
					WillReturnResult(sqlmock.NewResult(1, 1))
			},
			wantErr: false,
		},
		{
			name: "Database error",
			pvz: &models.PVZ{
				ID:               "test-id",
				RegistrationDate: time.Now(),
				City:             "Москва",
			},
			mock: func() {
				mock.ExpectExec("INSERT INTO pvz").
					WithArgs("test-id", sqlmock.AnyArg(), "Москва").
					WillReturnError(sql.ErrConnDone)
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.mock()
			err := repo.CreatePVZ(tt.pvz)
			if tt.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
			assert.NoError(t, mock.ExpectationsWereMet())
		})
	}
}

func TestPVZRepository_ListPVZ(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("an error '%s' was not expected when opening a stub database connection", err)
	}
	defer db.Close()

	repo := NewPVZRepository(db)

	now := time.Now()
	startDate := now.Add(-24 * time.Hour)
	endDate := now.Add(24 * time.Hour)

	tests := []struct {
		name      string
		limit     int
		offset    int
		startDate *time.Time
		endDate   *time.Time
		mock      func()
		wantErr   bool
	}{
		{
			name:      "List with pagination",
			limit:     10,
			offset:    0,
			startDate: nil,
			endDate:   nil,
			mock: func() {
				rows := sqlmock.NewRows([]string{"id", "registrationDate", "city"}).
					AddRow("1", now, "Москва").
					AddRow("2", now, "Санкт-Петербург")
				mock.ExpectQuery("SELECT id, registrationDate, city FROM pvz").
					WillReturnRows(rows)
			},
			wantErr: false,
		},
		{
			name:      "List with date range",
			limit:     10,
			offset:    0,
			startDate: &startDate,
			endDate:   &endDate,
			mock: func() {
				rows := sqlmock.NewRows([]string{"id", "registrationDate", "city"}).
					AddRow("1", now, "Москва")
				mock.ExpectQuery("SELECT id, registrationDate, city FROM pvz").
					WithArgs(startDate, endDate).
					WillReturnRows(rows)
			},
			wantErr: false,
		},
		{
			name:      "Database error",
			limit:     10,
			offset:    0,
			startDate: nil,
			endDate:   nil,
			mock: func() {
				mock.ExpectQuery("SELECT id, registrationDate, city FROM pvz").
					WillReturnError(sql.ErrConnDone)
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.mock()
			pvzs, err := repo.ListPVZ(tt.limit, tt.offset, tt.startDate, tt.endDate)
			if tt.wantErr {
				assert.Error(t, err)
				assert.Nil(t, pvzs)
			} else {
				assert.NoError(t, err)
				assert.NotNil(t, pvzs)
			}
			assert.NoError(t, mock.ExpectationsWereMet())
		})
	}
}
