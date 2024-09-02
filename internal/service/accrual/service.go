package accrual

import (
	"github.com/AxMdv/go-gophermart/internal/service/reward"
	"github.com/AxMdv/go-gophermart/internal/storage"
)

type AccrualService struct {
	repository  *storage.DBRepository
	RewardQueue *reward.Queue
}

func New(dbRepository *storage.DBRepository, queue *reward.Queue) *AccrualService {
	return &AccrualService{repository: dbRepository, RewardQueue: queue}
}
