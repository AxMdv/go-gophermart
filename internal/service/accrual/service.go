package accrual

import (
	"context"
	"errors"
	"log"
	"time"

	"github.com/AxMdv/go-gophermart/internal/model"
	"github.com/AxMdv/go-gophermart/internal/storage"
)

type AccrualService struct {
	Queue      *Queue
	requester  *Requester
	repository *storage.DBRepository
}

func NewService(accrualSystemAddr string, repo *storage.DBRepository) *AccrualService {
	queue := NewQueue()

	requester := NewRequester(accrualSystemAddr)

	a := AccrualService{
		Queue:      queue,
		requester:  requester,
		repository: repo,
	}

	go a.Loop()

	return &a
}

type Task struct {
	Order *model.Order
	Addr  string
}

func (a *AccrualService) Loop() {
	for {
		t, found := a.Queue.PopWait()
		if !found {
			continue
		}

		resp, err := a.requester.RewardRequest(t.Order, t.Addr)
		ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
		if err != nil {
			if errors.Is(err, ErrOrderNotRegistered) {
				order := &model.Order{
					ID:       t.Order.ID,
					UserUUID: t.Order.UserUUID,
					Accrual:  0,
					Status:   model.OrderStatusInvalid,
				}
				err = a.repository.UpdateOrder(ctx, order)
				if err != nil {
					log.Printf("error: %v\n", err)
					a.Queue.RemoveLastCompleted()
					continue
				}

			}
			log.Printf("error: %v\n", err)
			a.Queue.RemoveLastCompleted()
			continue
		}
		order := &model.Order{
			ID:       t.Order.ID,
			UserUUID: t.Order.UserUUID,
			Accrual:  resp.Accrual,
			Status:   resp.Status,
		}
		err = a.repository.UpdateOrder(ctx, order)
		if err != nil {
			log.Printf("error: %v\n", err)
			a.Queue.RemoveLastCompleted()
			continue
		}
		err = a.repository.UpdateUserBalance(ctx, order)
		if err != nil {
			log.Printf("error: %v\n", err)
			a.Queue.RemoveLastCompleted()
			continue
		}
		cancel()
		a.Queue.RemoveLastCompleted()
		log.Printf("accrual worker done request %v %v\n", t.Order, order)
	}
}

var ErrOrderNotRegistered = errors.New("204")
var ErrTooManyRequests = errors.New("429 should sleep all goroutines")
