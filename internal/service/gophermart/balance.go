package gophermart

import (
	"context"

	"github.com/AxMdv/go-gophermart/internal/model"
)

func (g *GophermartService) GetUserBalance(ctx context.Context, userID string) (*model.Balance, error) {
	balance, err := g.repository.GetUserBalance(ctx, userID)
	return balance, err
}
