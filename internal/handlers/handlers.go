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
	"github.com/AxMdv/go-gophermart/internal/service/gophermart"

	"github.com/AxMdv/go-gophermart/internal/storage"
)

type Handlers struct {
	gophermartService *gophermart.GophermartService
	accrualService    *accrual.AccrualService
	config            *config.Config
}

func New(a *gophermart.GophermartService, c *config.Config) *Handlers {
	return &Handlers{gophermartService: a, config: c}
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

	uuid, err := h.gophermartService.RegisterUser(ctx, &user)
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

	userID, err := h.gophermartService.LoginUser(ctx, &user)
	if err != nil {
		if errors.Is(err, gophermart.ErrInvalidAuthData) {
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

	valid := h.gophermartService.ValidateOrderID(orderID)
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

	err = h.gophermartService.CreateOrder(ctx, order)
	if err != nil {
		if errors.Is(err, gophermart.ErrOrderCreatedByCurrentUser) {
			w.WriteHeader(http.StatusOK)
			return
		}
		if errors.Is(err, gophermart.ErrOrderCreatedByAnotherUser) {
			w.WriteHeader(http.StatusConflict)
			return
		}
		log.Println(err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	task := &accrual.Task{
		Order: order,
		Addr:  h.config.AccrualSystemAddr,
	}
	h.accrualService.Queue.Push(task)
	fmt.Println(err)
	w.WriteHeader(http.StatusAccepted)
}

func (h *Handlers) GetOrdersInfo(w http.ResponseWriter, r *http.Request) {
	id := auth.GetUUIDFromContext(r.Context())

	ctx, cancel := context.WithTimeout(context.Background(), 900*time.Millisecond)
	defer cancel()
	orders, err := h.gophermartService.GetOrdersByUserID(ctx, id)
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
	log.Println("we response orders ..", orders)
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	w.Write(resp)
}

func (h *Handlers) GetWithdrawalsInfo(w http.ResponseWriter, r *http.Request) {
	userID := auth.GetUUIDFromContext(r.Context())
	ctx, cancel := context.WithTimeout(context.Background(), 120*time.Second)
	defer cancel()

	withdrawals, err := h.gophermartService.GetWithdrawalsInfo(ctx, userID)
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
	balance, err := h.gophermartService.GetUserBalance(ctx, id)
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

	valid := h.gophermartService.ValidateOrderID(orderID)
	if !valid {
		log.Println(err)
		w.WriteHeader(http.StatusUnprocessableEntity)
		return
	}

	withdrawal.UserUUID = auth.GetUUIDFromContext(r.Context())

	ctx, cancel := context.WithTimeout(context.Background(), 120*time.Second)
	defer cancel()

	err = h.gophermartService.CreateWithdraw(ctx, &withdrawal)
	if err != nil {
		if errors.Is(err, gophermart.ErrLowBalance) {
			w.WriteHeader(http.StatusPaymentRequired)
			return
		}
		log.Println(err)
		w.WriteHeader(http.StatusInternalServerError)
		return

	}
	w.WriteHeader(http.StatusOK)
}
