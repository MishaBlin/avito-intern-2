package repository

import (
	"avito-intern/internal/api/dto/internalErrors"
	"avito-intern/internal/models"
	"database/sql"
	"errors"

	"github.com/Masterminds/squirrel"
)

type UserRepositoryInterface interface {
	CreateUser(user *models.User) error
	GetUserByEmail(email string) (*models.User, error)
}

type UserRepository struct {
	db         *sql.DB
	sqlBuilder squirrel.StatementBuilderType
}

func NewUserRepository(db *sql.DB) *UserRepository {
	return &UserRepository{
		db:         db,
		sqlBuilder: squirrel.StatementBuilder.PlaceholderFormat(squirrel.Dollar),
	}
}

func (r *UserRepository) CreateUser(user *models.User) error {
	query, args, err := r.sqlBuilder.
		Insert("users").
		Columns("id", "email", "password", "role").
		Values(user.ID, user.Email, user.Password, user.Role).
		ToSql()
	if err != nil {
		return err
	}
	_, err = r.db.Exec(query, args...)
	return err
}

func (r *UserRepository) GetUserByEmail(email string) (*models.User, error) {
	var user models.User
	query, args, err := r.sqlBuilder.
		Select("id", "email", "password", "role").
		From("users").
		Where(squirrel.Eq{"email": email}).
		ToSql()
	if err != nil {
		return nil, err
	}

	err = r.db.QueryRow(query, args...).Scan(&user.ID, &user.Email, &user.Password, &user.Role)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, internalErrors.ErrUserNotFound
		}
		return nil, err
	}
	return &user, nil
}
