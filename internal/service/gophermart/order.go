package gophermart

import (
	"context"
	"errors"
	"time"

	"github.com/AxMdv/go-gophermart/internal/model"
	"github.com/AxMdv/go-gophermart/internal/storage"
	"github.com/theplant/luhn"
)

func (g *GophermartService) ValidateOrderID(orderID int) (valid bool) {

	valid = luhn.Valid(orderID)
	return valid
}

func (g *GophermartService) CreateOrder(ctx context.Context, order *model.Order) (err error) {
	// t := time.Now()
	// order.UploadedAt = t.Format(time.RFC3339)
	order.UploadedAt = time.Now()
	id, err := g.repository.GetOrderByID(ctx, order)
	if err != nil {
		if errors.Is(err, storage.ErrNoOrder) {
			err = g.repository.CreateOrder(ctx, order)
			return err
		}
		return err
	}
	if id == order.UserUUID {
		return ErrOrderCreatedByCurrentUser
	}
	return ErrOrderCreatedByAnotherUser
}

func (g *GophermartService) GetOrdersByUserID(ctx context.Context, userID string) (orders []model.Order, err error) {
	orders, err = g.repository.GetOrdersByUserID(ctx, userID)
	return orders, err
}

var ErrOrderCreatedByCurrentUser = errors.New("order created by current user")
var ErrOrderCreatedByAnotherUser = errors.New("order created by another user")
