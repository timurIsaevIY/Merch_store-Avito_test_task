package repository

import (
	"Merch_store-Avito_test_task/internal/models"
	"context"
	"database/sql"
	"errors"
	"github.com/lib/pq"
)

type AuthRepositoryImpl struct {
	db *sql.DB
}

func NewAuthRepositoryImpl(db *sql.DB) *AuthRepositoryImpl {
	return &AuthRepositoryImpl{db: db}
}

func (repo *AuthRepositoryImpl) CreateUser(ctx context.Context, user models.User) (uint, error) {
	var userID uint
	query := `INSERT INTO "user" (username, password_hash) VALUES ($1, $2) RETURNING id`
	err := repo.db.QueryRowContext(ctx, query, user.Username, user.PasswordHash).Scan(&userID)
	if err != nil {
		var pqErr *pq.Error
		if errors.As(err, &pqErr) && pqErr.Code == "23505" {
			return 0, models.ErrAlreadyExists
		}
		return 0, err
	}
	return userID, nil
}

func (repo *AuthRepositoryImpl) GetUser(ctx context.Context, username string) (models.User, error) {
	query := `SELECT id, username, password_hash FROM "user" WHERE username = $1`
	row := repo.db.QueryRowContext(ctx, query, username)
	var user models.User
	err := row.Scan(&user.ID, &user.Username, &user.PasswordHash)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return models.User{}, models.ErrNotFound
		}
		return models.User{}, err
	}
	return user, nil
}
