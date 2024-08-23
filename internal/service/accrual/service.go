package accrual

import (
	"github.com/AxMdv/go-gophermart/internal/storage"
)

type AccrualService struct {
	repository *storage.DBRepository
}

func New(dbRepository *storage.DBRepository) *AccrualService {
	return &AccrualService{repository: dbRepository}
}
