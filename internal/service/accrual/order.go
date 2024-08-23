package accrual

import (
	"context"

	"github.com/AxMdv/go-gophermart/internal/model"
	"github.com/theplant/luhn"
)

func (a *AccrualService) ValidateOrderID(orderID int) (valid bool) {

	valid = luhn.Valid(orderID)
	return valid
}

func (a *AccrualService) CreateOrder(ctx context.Context, order *model.Order) (err error) {

	return err
}
