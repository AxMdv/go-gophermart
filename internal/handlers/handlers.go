package handlers

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"log"
	"net/http"
	"strconv"
	"time"

	"github.com/AxMdv/go-gophermart/internal/config"
	"github.com/AxMdv/go-gophermart/internal/model"
	"github.com/AxMdv/go-gophermart/internal/service/accrual"
	"github.com/AxMdv/go-gophermart/internal/service/auth"
	"github.com/AxMdv/go-gophermart/internal/service/reward"
	"github.com/AxMdv/go-gophermart/internal/storage"
)

type Handlers struct {
	accrualService *accrual.AccrualService
	config         *config.Options
}

func New(a *accrual.AccrualService, c *config.Options) *Handlers {
	return &Handlers{accrualService: a, config: c}
}

func (h *Handlers) RegisterUser(w http.ResponseWriter, r *http.Request) {
	if r.Header.Get("Content-Type") != "application/json" {
		w.WriteHeader(http.StatusBadRequest)
		return
	}
	var user model.User
	bodyBytes, err := io.ReadAll(r.Body)
	if err != nil {
		log.Println(err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	err = json.Unmarshal(bodyBytes, &user)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		return
	}
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	uuid, err := h.accrualService.RegisterUser(ctx, &user)
	if err != nil {
		if errors.Is(err, storage.ErrLoginDuplicate) {
			w.WriteHeader(http.StatusConflict)
			return
		}
		log.Println(err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	cookie := auth.CreateCookie(uuid)

	http.SetCookie(w, cookie)
	w.WriteHeader(http.StatusOK)
}

func (h *Handlers) LoginUser(w http.ResponseWriter, r *http.Request) {
	if r.Header.Get("Content-Type") != "application/json" {
		w.WriteHeader(http.StatusBadRequest)
		return
	}
	var user model.User
	bodyBytes, err := io.ReadAll(r.Body)
	if err != nil {
		log.Println(err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	err = json.Unmarshal(bodyBytes, &user)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	userID, err := h.accrualService.LoginUser(ctx, &user)
	if err != nil {
		if errors.Is(err, accrual.ErrInvalidAuthData) {
			w.WriteHeader(http.StatusUnauthorized)
			return
		}
		w.WriteHeader(http.StatusInternalServerError)
		log.Println(err)
		return
	}
	cookie := auth.CreateCookie(userID)

	http.SetCookie(w, cookie)
	w.WriteHeader(http.StatusOK)
}

func (h *Handlers) CreateOrder(w http.ResponseWriter, r *http.Request) {
	if r.Header.Get("Content-Type") != "text/plain" {
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	bodyBytes, err := io.ReadAll(r.Body)
	if err != nil {
		log.Println(err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	orderID, err := strconv.Atoi(string(bodyBytes))
	if err != nil {
		w.WriteHeader(http.StatusUnprocessableEntity)
		return
	}

	valid := h.accrualService.ValidateOrderID(orderID)
	if !valid {
		w.WriteHeader(http.StatusUnprocessableEntity)
		return
	}

	id := auth.GetUUIDFromContext(r.Context())

	ctx, cancel := context.WithTimeout(context.Background(), 120*time.Second)
	defer cancel()

	order := &model.Order{
		UserUUID: id,
		ID:       strconv.Itoa(orderID),
		Status:   model.OrderStatusNew,
	}

	err = h.accrualService.CreateOrder(ctx, order)
	if err != nil {
		if errors.Is(err, accrual.ErrOrderCreatedByCurrentUser) {
			w.WriteHeader(http.StatusOK)
			return
		}
		if errors.Is(err, accrual.ErrOrderCreatedByAnotherUser) {
			w.WriteHeader(http.StatusConflict)
			return
		}
		log.Println(err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	// accrual.InputCh <- order

	// err = h.accrualService.RewardRequest(order, h.config.AccrualSystemAddr+"/api/orders/"+order.ID)

	task := &reward.Task{
		Order: order,
		Addr:  h.config.AccrualSystemAddr,
	}
	h.accrualService.RewardQueue.Push(task)
	fmt.Println(err)
	w.WriteHeader(http.StatusAccepted)
}

func (h *Handlers) GetOrdersInfo(w http.ResponseWriter, r *http.Request) {
	id := auth.GetUUIDFromContext(r.Context())

	ctx, cancel := context.WithTimeout(context.Background(), 900*time.Millisecond)
	defer cancel()
	orders, err := h.accrualService.GetOrdersByUserID(ctx, id)
	if err != nil {
		if errors.Is(err, storage.ErrNoOrders) {
			w.WriteHeader(http.StatusNoContent)
			return
		}
		log.Println(err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	resp, err := json.Marshal(orders)
	if err != nil {
		log.Println(err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	w.Write(resp)
}

func (h *Handlers) GetWithdrawalsInfo(w http.ResponseWriter, r *http.Request) {
	userID := auth.GetUUIDFromContext(r.Context())
	ctx, cancel := context.WithTimeout(context.Background(), 120*time.Second)
	defer cancel()

	withdrawals, err := h.accrualService.GetWithdrawalsInfo(ctx, userID)
	if err != nil {
		if errors.Is(err, storage.ErrNoWithdrawalsData) {
			w.WriteHeader(http.StatusNoContent)
			return
		}
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	resp, err := json.Marshal(withdrawals)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	w.Write(resp)
}

func (h *Handlers) GetUserBalance(w http.ResponseWriter, r *http.Request) {

	id := auth.GetUUIDFromContext(r.Context())
	ctx, cancel := context.WithTimeout(context.Background(), 120*time.Second)
	defer cancel()
	balance, err := h.accrualService.GetUserBalance(ctx, id)
	if err != nil {
		log.Println(err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	resp, err := json.Marshal(balance)
	if err != nil {
		log.Println(err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	w.Write(resp)
}

func (h *Handlers) CreateWithdraw(w http.ResponseWriter, r *http.Request) {
	if r.Header.Get("Content-Type") != "application/json" {
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	bodyBytes, err := io.ReadAll(r.Body)
	if err != nil {
		log.Println(err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	var withdrawal model.Withdrawal
	err = json.Unmarshal(bodyBytes, &withdrawal)
	if err != nil {
		log.Println(err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	orderID, err := strconv.Atoi(withdrawal.OrderID)
	if err != nil {
		log.Println(err)
		w.WriteHeader(http.StatusUnprocessableEntity)
		return
	}

	valid := h.accrualService.ValidateOrderID(orderID)
	if !valid {
		log.Println(err)
		w.WriteHeader(http.StatusUnprocessableEntity)
		return
	}

	withdrawal.UserUUID = auth.GetUUIDFromContext(r.Context())

	ctx, cancel := context.WithTimeout(context.Background(), 120*time.Second)
	defer cancel()

	err = h.accrualService.CreateWithdraw(ctx, &withdrawal)
	if err != nil {
		if errors.Is(err, accrual.ErrLowBalance) {
			w.WriteHeader(http.StatusPaymentRequired)
			return
		}
		log.Println(err)
		w.WriteHeader(http.StatusInternalServerError)
		return

	}
	w.WriteHeader(http.StatusOK)
}
