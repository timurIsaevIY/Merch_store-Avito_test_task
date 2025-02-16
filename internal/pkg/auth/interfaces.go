package auth

import (
	"Merch_store-Avito_test_task/internal/models"
	"context"
)

//go:generate mockgen -source=interfaces.go -destination=mocks/mock.go
type AuthUsecase interface {
	Login(ctx context.Context, username, password string) (models.User, error)
}

type AuthRepository interface {
	CreateUser(ctx context.Context, user models.User) (uint, error)
	GetUser(ctx context.Context, username string) (models.User, error)
}
