package repository

import (
	"avito-intern/internal/api/dto/internalErrors"
	"avito-intern/internal/models"
	"database/sql"
	"testing"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/stretchr/testify/assert"
)

func TestUserRepository_CreateUser_Success(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("an error '%s' was not expected when opening a stub database connection", err)
	}
	defer db.Close()

	repo := NewUserRepository(db)

	user := &models.User{
		ID:       "test-id",
		Email:    "test@example.com",
		Password: "hashed_password",
		Role:     "moderator",
	}

	mock.ExpectExec("INSERT INTO users").
		WithArgs("test-id", "test@example.com", "hashed_password", "moderator").
		WillReturnResult(sqlmock.NewResult(1, 1))

	err = repo.CreateUser(user)

	assert.NoError(t, err)
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestUserRepository_CreateUser_DatabaseError(t *testing.T) {

	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("an error '%s' was not expected when opening a stub database connection", err)
	}
	defer db.Close()

	repo := NewUserRepository(db)

	user := &models.User{
		ID:       "test-id",
		Email:    "test@example.com",
		Password: "hashed_password",
		Role:     "moderator",
	}

	mock.ExpectExec("INSERT INTO users").
		WithArgs("test-id", "test@example.com", "hashed_password", "moderator").
		WillReturnError(sql.ErrConnDone)

	err = repo.CreateUser(user)

	assert.Error(t, err)
	assert.Equal(t, sql.ErrConnDone, err)
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestUserRepository_GetUserByEmail_Success(t *testing.T) {

	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("an error '%s' was not expected when opening a stub database connection", err)
	}
	defer db.Close()

	repo := NewUserRepository(db)

	rows := sqlmock.NewRows([]string{"id", "email", "password", "role"}).
		AddRow("test-id", "test@example.com", "hashed_password", "moderator")
	mock.ExpectQuery("SELECT id, email, password, role FROM users").
		WithArgs("test@example.com").
		WillReturnRows(rows)

	user, err := repo.GetUserByEmail("test@example.com")

	assert.NoError(t, err)
	assert.NotNil(t, user)
	assert.Equal(t, "test-id", user.ID)
	assert.Equal(t, "test@example.com", user.Email)
	assert.Equal(t, "hashed_password", user.Password)
	assert.Equal(t, "moderator", user.Role)
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestUserRepository_GetUserByEmail_NotFound(t *testing.T) {

	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("an error '%s' was not expected when opening a stub database connection", err)
	}
	defer db.Close()

	repo := NewUserRepository(db)

	mock.ExpectQuery("SELECT id, email, password, role FROM users").
		WithArgs("nonexistent@example.com").
		WillReturnError(sql.ErrNoRows)

	user, err := repo.GetUserByEmail("nonexistent@example.com")

	assert.Error(t, err)
	assert.Equal(t, internalErrors.ErrUserNotFound, err)
	assert.Nil(t, user)
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestUserRepository_GetUserByEmail_DatabaseError(t *testing.T) {

	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("an error '%s' was not expected when opening a stub database connection", err)
	}
	defer db.Close()

	repo := NewUserRepository(db)

	mock.ExpectQuery("SELECT id, email, password, role FROM users").
		WithArgs("test@example.com").
		WillReturnError(sql.ErrConnDone)

	user, err := repo.GetUserByEmail("test@example.com")

	assert.Error(t, err)
	assert.Equal(t, sql.ErrConnDone, err)
	assert.Nil(t, user)
	assert.NoError(t, mock.ExpectationsWereMet())
}
