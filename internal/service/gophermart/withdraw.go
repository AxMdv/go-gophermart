package gophermart

import (
	"context"
	"errors"
	"time"

	"github.com/AxMdv/go-gophermart/internal/model"
)

func (g *GophermartService) GetWithdrawalsInfo(ctx context.Context, userID string) ([]model.Withdrawal, error) {
	withdrawals, err := g.repository.GetWithdrawalsByUserID(ctx, userID)
	return withdrawals, err
}

func (g *GophermartService) CreateWithdraw(ctx context.Context, withdrawal *model.Withdrawal) error {
	userBalance, err := g.repository.GetUserBalance(ctx, withdrawal.UserUUID)
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
	err = g.repository.CreateWithdraw(ctx, userBalance, withdrawal)
	return err
}

var ErrLowBalance = errors.New("user balance is too low")
