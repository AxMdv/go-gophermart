package accrual

import (
	"context"
	"errors"

	"github.com/AxMdv/go-gophermart/internal/model"
	"github.com/AxMdv/go-gophermart/internal/storage"
)

func (a *AccrualService) LoginUser(ctx context.Context, user *model.User) (userID string, err error) {
	reqUser := user
	dbUser, err := a.repository.GetUserAuthData(ctx, user)
	if err != nil {
		if errors.Is(err, storage.ErrNoAuthData) {
			return "", ErrInvalidAuthData
		}
		return "", err
	}
	if reqUser.Login == dbUser.Login && reqUser.Password == dbUser.Password {
		return dbUser.UUID, nil
	}
	return "", ErrInvalidAuthData
}

var ErrInvalidAuthData = errors.New("invalid username and password")
