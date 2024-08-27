package accrual

import (
	"context"

	"github.com/AxMdv/go-gophermart/internal/model"
)

func (a *AccrualService) GetUserBalance(ctx context.Context, userID string) (*model.Balance, error) {
	balance, err := a.repository.GetUserBalance(ctx, userID)
	return balance, err
}
