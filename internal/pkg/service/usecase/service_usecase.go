package service

import (
	"Merch_store-Avito_test_task/internal/models"
	"Merch_store-Avito_test_task/internal/pkg/middleware"
	"Merch_store-Avito_test_task/internal/pkg/service"
	"context"
)

type ServiceUsecaseImpl struct {
	repo service.ServiceRepository
}

func NewServiceUsecase(repo service.ServiceRepository) *ServiceUsecaseImpl {
	return &ServiceUsecaseImpl{repo}
}

func (u *ServiceUsecaseImpl) GetUserInfo(ctx context.Context) (models.UserData, error) {
	userID := ctx.Value(middleware.IdKey).(uint)
	return u.repo.GetUserInfo(ctx, userID)
}
