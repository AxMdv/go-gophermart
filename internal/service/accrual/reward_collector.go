package accrual

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"log"
	"net/http"

	"github.com/AxMdv/go-gophermart/internal/model"
)

//	func asd() {
//		inputCh := make(chan model.Order)
//	}
type RewardResponse struct {
	Order   string  `json:"order"`
	Status  string  `json:"status"`
	Accrual float32 `json:"accrual"`
}

func (a *AccrualService) RewardRequest(order *model.Order, addr string) error {
	// url := addr + "/api/orders/"
	req, err := http.NewRequest(http.MethodGet, addr, nil)
	if err != nil {
		log.Println(err)
		return err
	}

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		log.Println(err)
		return err
	}
	switch resp.StatusCode {
	case 200:
		bytes, err := io.ReadAll(resp.Body)
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
	case 204:
		return ErrOrderNotRegistered
	case 429:
		return ErrTooManyRequests
	default:
		fmt.Println("default")
	}
	return nil
}

var ErrOrderNotRegistered = errors.New("204")
var ErrTooManyRequests = errors.New("429 should sleep all goroutines")
