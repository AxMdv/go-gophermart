package accrual

import (
	"context"
	"errors"
	"time"

	"github.com/AxMdv/go-gophermart/internal/model"
)

func (a *AccrualService) GetWithdrawalsInfo(ctx context.Context, userID string) ([]model.Withdrawal, error) {
	withdrawals, err := a.repository.GetWithdrawalsByUserID(ctx, userID)
	return withdrawals, err
}

func (a *AccrualService) CreateWithdraw(ctx context.Context, withdrawal *model.Withdrawal) error {
	userBalance, err := a.repository.GetUserBalance(ctx, withdrawal.UserUUID)
	if err != nil {
		return err
	}

	if withdrawal.Amount > userBalance.Current {
		return ErrLowBalance
	}
	userBalance.Current -= withdrawal.Amount
	userBalance.Withdrawn += withdrawal.Amount
	userBalance.UserUUID = withdrawal.UserUUID
	withdrawal.ProcessedAt = time.Now()
	err = a.repository.CreateWithdraw(ctx, userBalance, withdrawal)
	return err
}

var ErrLowBalance = errors.New("user balance is too low")
