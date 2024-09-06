package accrual

import (
	"encoding/json"
	"errors"
	"io"
	"log"
	"net/http"
	"strconv"
	"time"

	"github.com/AxMdv/go-gophermart/internal/model"
)

type RewardResponse struct {
	Order   string  `json:"order"`
	Status  string  `json:"status"`
	Accrual float64 `json:"accrual"`
}

type Requester struct {
	accrualAddr string
}

func NewRequester(addr string) *Requester {
	return &Requester{
		accrualAddr: addr,
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

	for {
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
			secondsToSleepStr := resp.Header.Get("Retry-After")
			if secondsToSleepStr == "" {
				time.Sleep(60 * time.Second)
				continue
			}
			secondsToSleep, err := strconv.Atoi(secondsToSleepStr)
			if err != nil {
				time.Sleep(60 * time.Second)
				continue
			}
			time.Sleep(time.Duration(secondsToSleep) * time.Second)
		default:
			log.Println("default branch .. status code is ", resp.StatusCode)
			return rr, errors.New("unexpected error")
		}
	}
}
