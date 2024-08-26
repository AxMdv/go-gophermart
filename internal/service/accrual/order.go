package accrual

import (
	"context"
	"errors"
	"time"

	"github.com/AxMdv/go-gophermart/internal/model"
	"github.com/AxMdv/go-gophermart/internal/storage"
	"github.com/theplant/luhn"
)

func (a *AccrualService) ValidateOrderID(orderID int) (valid bool) {

	valid = luhn.Valid(orderID)
	return valid
}

func (a *AccrualService) CreateOrder(ctx context.Context, order *model.Order) (err error) {
	// t := time.Now()
	// order.UploadedAt = t.Format(time.RFC3339)
	order.UploadedAt = time.Now()
	id, err := a.repository.GetOrderByID(ctx, order)
	if err != nil {
		if errors.Is(err, storage.ErrNoOrder) {
			err = a.repository.CreateOrder(ctx, order)
			return err
		}
		return err
	}
	if id == order.UserUUID {
		return ErrOrderCreatedByCurrentUser
	}
	return ErrOrderCreatedByAnotherUser
}

func (a *AccrualService) GetOrdersByUserID(ctx context.Context, userID string) (orders []model.Order, err error) {
	orders, err = a.repository.GetOrdersByUserID(ctx, userID)
	return orders, err
}

var ErrOrderCreatedByCurrentUser = errors.New("order created by current user")
var ErrOrderCreatedByAnotherUser = errors.New("order created by another user")
