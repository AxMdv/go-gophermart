package accrual

import (
	"github.com/AxMdv/go-gophermart/internal/storage"
)

type AccrualService struct {
	repository  *storage.DBRepository
	RewardQueue *Queue
}

func New(dbRepository *storage.DBRepository, queue *Queue) *AccrualService {
	return &AccrualService{repository: dbRepository, RewardQueue: queue}
}
