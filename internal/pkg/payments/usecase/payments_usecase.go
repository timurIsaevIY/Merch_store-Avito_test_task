package usecase

import (
	"Merch_store-Avito_test_task/internal/pkg/payments"
	"context"
)

type PaymentsUsecaseImpl struct {
	repo payments.PaymentsRepository
}

func NewPaymentsUsecase(repo payments.PaymentsRepository) *PaymentsUsecaseImpl {
	return &PaymentsUsecaseImpl{repo}
}

func (r *PaymentsUsecaseImpl) SendCoins(ctx context.Context, toUser string, amount uint) error {
	return r.repo.Transfer(ctx, toUser, amount)
}

func (r *PaymentsUsecaseImpl) BuyItem(ctx context.Context, itemId uint) error {
	return r.repo.BuyItem(ctx, itemId)
}
