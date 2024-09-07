package gophermart

import (
	"context"

	"github.com/AxMdv/go-gophermart/internal/model"
	"github.com/google/uuid"
)

func (g *GophermartService) RegisterUser(ctx context.Context, user *model.User) (newUUID string, err error) {
	id, err := createUUID()
	if err != nil {
		return "", err
	}
	newUUID = id.String()
	user.UUID = newUUID

	err = g.repository.RegisterUser(ctx, user)
	return newUUID, err
}

func createUUID() (uuid.UUID, error) {
	newID, err := uuid.NewV6()
	if err != nil {
		return newID, err
	}
	return newID, nil
}
