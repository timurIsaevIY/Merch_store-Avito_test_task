package payments

import "context"

//go:generate mockgen -source=interfaces.go -destination=mocks/mock.go
type PaymentsUsecase interface {
	SendCoins(ctx context.Context, toUser string, amount uint) error
	BuyItem(ctx context.Context, itemId uint) error
}

type PaymentsRepository interface {
	Transfer(ctx context.Context, toUser string, amount uint) error
	BuyItem(ctx context.Context, itemId uint) error
}
