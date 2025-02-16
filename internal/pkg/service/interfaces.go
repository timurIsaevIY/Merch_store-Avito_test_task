package service

import (
	"Merch_store-Avito_test_task/internal/models"
	"context"
)

//go:generate mockgen -source=interfaces.go -destination=mocks/mock.go
type ServiceUsecase interface {
	GetUserInfo(ctx context.Context) (models.UserData, error)
}

type ServiceRepository interface {
	GetUserInfo(ctx context.Context, userID uint) (models.UserData, error)
}
