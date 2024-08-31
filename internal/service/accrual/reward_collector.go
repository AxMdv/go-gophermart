package accrual

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"log"
	"net/http"
	"time"

	"github.com/AxMdv/go-gophermart/internal/model"
)

type RewardResponse struct {
	Order   string  `json:"order"`
	Status  string  `json:"status"`
	Accrual float32 `json:"accrual"`
}

// var InputCh chan *model.Order

var numberWorkers = 10

// func Init() {
// 	var InputCh = make(chan *model.Order, 100)
// 	defer close(InputCh)
// }

type Task struct {
	Order *model.Order
	Addr  string
}

// type RewardCollector struct {
// 	queue Queue
// }

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
	idk string
}

func NewRequester() *Requester {
	return &Requester{
		idk: "123",
	}
}

type Worker struct {
	id        int
	queue     *Queue
	requester *Requester
}

func NewWorker(id int, queue *Queue, requester *Requester) *Worker {
	w := Worker{
		id:        id,
		queue:     queue,
		requester: requester,
	}
	return &w
}

func (w *Worker) Loop() {
	for {
		t := w.queue.PopWait()

		err := w.requester.RewardRequest(t.Order, t.Addr)
		if err != nil {
			fmt.Printf("error: %v\n", err)
			continue
		}

		fmt.Printf("worker #%d done request %v\n", w.id, t.Order)
	}
}

func (r *Requester) RewardRequest(order *model.Order, addr string) error {
	// url := addr + "/api/orders/"
	req, err := http.NewRequest(http.MethodGet, addr, nil)
	if err != nil {
		log.Println(err)
		return err
	}

	client := &http.Client{}

	loop := true

	for loop {
		resp, err := client.Do(req)
		if err != nil {
			log.Println(err)
			return err
		}

		switch resp.StatusCode {
		case 200:
			bytes, err := io.ReadAll(resp.Body)
			defer resp.Body.Close()
			if err != nil {
				log.Println(err)
				return err
			}
			var rr RewardResponse
			err = json.Unmarshal(bytes, &rr)
			if err != nil {
				log.Println(err)
				return err
			}
			switch rr.Status {
			case model.OrderStatusRegistered:
				fmt.Println("retry")
				continue
			case model.OrderStatusInvalid:
				return ErrOrderNotRegistered
			case model.OrderStatusProcessing:

				fmt.Println("retry")
				continue
			case model.OrderStatusProcessed:
				fmt.Println(rr.Status)
				continue
			default:
				fmt.Println(rr.Status)
				continue
			}

		case 204:
			return ErrOrderNotRegistered
		case 429:
			time.Sleep(60 * time.Second)
		default:
			fmt.Println("default branch .. status code is ", resp.StatusCode)
			return errors.New("unexpected error")
		}
	}
	return err
}

func NewRewardCollectionProcess() *Queue {
	queue := NewQueue()

	for i := 0; i < numberWorkers; i++ {
		w := NewWorker(i, queue, NewRequester())
		go w.Loop()
	}
	return queue
}

var ErrOrderNotRegistered = errors.New("204")
var ErrTooManyRequests = errors.New("429 should sleep all goroutines")
