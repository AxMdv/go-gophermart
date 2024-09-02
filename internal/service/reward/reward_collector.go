package reward

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
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
		// err := w.requester.RegisterOrder(t.Order.ID)
		// if err != nil {
		// 	log.Printf("error: %v\n", err)
		// 	continue
		// }
		resp, err := w.requester.RewardRequest(t.Order, t.Addr)
		if err != nil {
			log.Printf("error: %v\n", err)
			break
		}
		log.Println("resp from accrual is ", resp)
		ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
		defer cancel()
		order := &model.Order{
			UserUUID: t.Order.UserUUID,
			Accrual:  resp.Accrual,
			Status:   resp.Status,
		}
		err = w.repository.UpdateOrder(ctx, order)
		if err != nil {
			log.Printf("error: %v\n", err)
			break
		}
		err = w.repository.UpdateUserBalance(ctx, order)
		if err != nil {
			log.Printf("error: %v\n", err)
			break
		}

		log.Printf("worker #%d done request %v %v\n", w.id, t.Order, order)
	}
}

type RegisterOrders struct {
	OrderID string `json:"order"`
	Goods   []Good `json:"goods"`
}

type Good struct {
	Description string `json:"description"`
	Price       int    `json:"price"`
}

func (r *Requester) RegisterOrder(orderID string) error {

	addr := fmt.Sprintf("%s/api/orders", r.accrualAddr)
	// if !strings.HasPrefix(addr, "http://") {
	// 	addr = fmt.Sprintf("http://%s", addr)
	// }
	reqBody := &RegisterOrders{
		OrderID: orderID,
		Goods:   []Good{{Description: "Bork чайник", Price: 100}},
	}
	body, err := json.Marshal(reqBody)
	if err != nil {
		return err
	}
	rdr := bytes.NewReader(body)
	req, err := http.NewRequest(http.MethodPost, addr, rdr)
	if err != nil {
		return err
	}
	req.Header.Add("Content-Type", "application/json")
	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	log.Println("Попытка зарегать ", addr, reqBody, resp.StatusCode)
	return err
}

func (r *Requester) RewardRequest(order *model.Order, addr string) (*RewardResponse, error) {

	rr := &RewardResponse{}
	// if !strings.HasPrefix(addr, "http://") {
	// 	addr = fmt.Sprintf("http://%s", addr)
	// }
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
		log.Println("resp status code from accrual is ", resp.StatusCode)
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
	// err := RewardRegister(fmt.Sprintf("%s/api/goods", addr))
	// if err != nil {
	// 	return &Queue{}, err
	// }

	for i := 0; i < numberWorkers; i++ {
		w := NewWorker(i, queue, NewRequester(addr), repository)
		go w.Loop()
	}
	return queue, nil
}

type Match struct {
	Match      string `json:"match"`
	Reward     int    `json:"reward"`
	RewardType string `json:"reward_type"`
}

func RewardRegister(addr string) error {
	match := &Match{
		Match:      "Bork",
		Reward:     10,
		RewardType: "%",
	}
	body, err := json.Marshal(match)
	if err != nil {
		return err
	}
	rdr := bytes.NewReader(body)
	// if !strings.HasPrefix(addr, "http://") {
	// 	addr = fmt.Sprintf("http://%s", addr)
	// }
	req, err := http.NewRequest(http.MethodPost, addr, rdr)
	req.Header.Add("Content-Type", "application/json")
	if err != nil {
		return err
	}
	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	log.Println("Попытка зарегать 10 % вознаграждения", resp.StatusCode)
	return err
}

var ErrOrderNotRegistered = errors.New("204")
var ErrTooManyRequests = errors.New("429 should sleep all goroutines")
