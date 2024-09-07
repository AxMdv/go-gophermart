package storage

import (
	"errors"
)

var ErrLoginDuplicate = errors.New("login already exists")
var ErrNoAuthData = errors.New("no auth data")
var ErrOrderDuplicate = errors.New("order already created")
var ErrNoOrder = errors.New("no order with current id")
var ErrNoOrders = errors.New("no orders with current user id")
var ErrNoWithdrawalsData = errors.New("no withdrawals with current userID")
