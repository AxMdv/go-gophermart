package reward

import (
	"context"
	"encoding/json"
	"errors"
	"io"
	"log"
	"net/http"
	"time"

	"github.com/AxMdv/go-gophermart/internal/model"
	"github.com/AxMdv/go-gophermart/internal/storage"
)

type RewardResponse struct {
	Order   string  `json:"order"`
	Status  string  `json:"status"`
	Accrual float64 `json:"accrual"`
}

var numberWorkers = 10

type Task struct {
	Order *model.Order
	Addr  string
}

type Queue struct {
	ch chan *Task
}

func NewQueue() *Queue {
	return &Queue{
		ch: make(chan *Task, 10),
	}
}

func (q *Queue) Push(t *Task) {
	// добавляем задачу в очередь
	q.ch <- t
}

func (q *Queue) PopWait() *Task {
	// получаем задачу
	return <-q.ch
}

type Requester struct {
	accrualAddr string
}

func NewRequester(addr string) *Requester {
	return &Requester{
		accrualAddr: addr,
	}
}

type Worker struct {
	id         int
	queue      *Queue
	requester  *Requester
	repository *storage.DBRepository
}

func NewWorker(id int, queue *Queue, requester *Requester, repo *storage.DBRepository) *Worker {
	w := Worker{
		id:         id,
		queue:      queue,
		requester:  requester,
		repository: repo,
	}
	return &w
}

func (w *Worker) Loop() {
	for {
		t := w.queue.PopWait()

		resp, err := w.requester.RewardRequest(t.Order, t.Addr)
		if err != nil {
			log.Printf("error: %v\n", err)
			continue
		}
		ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
		order := &model.Order{
			ID:       t.Order.ID,
			UserUUID: t.Order.UserUUID,
			Accrual:  resp.Accrual,
			Status:   resp.Status,
		}
		err = w.repository.UpdateOrder(ctx, order)
		if err != nil {
			log.Printf("error: %v\n", err)
			continue
		}
		err = w.repository.UpdateUserBalance(ctx, order)
		if err != nil {
			log.Printf("error: %v\n", err)
			continue
		}
		cancel()
		log.Printf("worker #%d done request %v %v\n", w.id, t.Order, order)
	}
}

func (r *Requester) RewardRequest(order *model.Order, addr string) (*RewardResponse, error) {

	rr := &RewardResponse{}

	url := addr + "/api/orders/" + order.ID
	req, err := http.NewRequest(http.MethodGet, url, nil)
	if err != nil {
		log.Println(err)
		return rr, err
	}

	client := &http.Client{}

	loop := true

	for loop {
		resp, err := client.Do(req)
		if err != nil {
			log.Println(err)
			return rr, err
		}
		switch resp.StatusCode {
		case 200:
			bytes, err := io.ReadAll(resp.Body)
			defer resp.Body.Close()
			if err != nil {
				log.Println(err)
				return rr, err
			}
			var rr RewardResponse
			err = json.Unmarshal(bytes, &rr)
			if err != nil {
				log.Println(err)
				return &rr, err
			}

			switch rr.Status {
			case model.OrderStatusRegistered:
				log.Println("retry..")
				continue
			case model.OrderStatusInvalid:
				return &rr, ErrOrderNotRegistered
			case model.OrderStatusProcessing:
				log.Println("retry..")
				continue
			case model.OrderStatusProcessed:
				log.Println(rr.Status)

				return &rr, nil

			default:
				log.Println(rr.Status, "default")
				continue
			}
		case 204:
			return rr, ErrOrderNotRegistered
		case 429:
			time.Sleep(60 * time.Second)
		default:
			log.Println("default branch .. status code is ", resp.StatusCode)
			return rr, errors.New("unexpected error")
		}
	}
	return rr, err
}

func NewRewardCollectionProcess(addr string, repository *storage.DBRepository) (*Queue, error) {

	queue := NewQueue()

	for i := 0; i < numberWorkers; i++ {
		w := NewWorker(i, queue, NewRequester(addr), repository)
		go w.Loop()
	}
	return queue, nil
}

var ErrOrderNotRegistered = errors.New("204")
var ErrTooManyRequests = errors.New("429 should sleep all goroutines")
