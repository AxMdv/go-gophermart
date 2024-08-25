package handlers

import (
	"context"
	"encoding/json"
	"errors"
	"io"
	"log"
	"net/http"
	"strconv"
	"time"

	"github.com/AxMdv/go-gophermart/internal/model"
	"github.com/AxMdv/go-gophermart/internal/service/accrual"
	"github.com/AxMdv/go-gophermart/internal/service/auth"
	"github.com/AxMdv/go-gophermart/internal/storage"
)

type Handlers struct {
	accrualService *accrual.AccrualService
}

func New(a *accrual.AccrualService) *Handlers {
	return &Handlers{accrualService: a}
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

	id := auth.GetUUIDFromContext(r.Context())

	user.UUID = id

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	err = h.accrualService.LoginUser(ctx, &user)
	if err != nil {
		if errors.Is(err, accrual.ErrInvalidAuthData) {
			w.WriteHeader(http.StatusUnauthorized)
			return
		}
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	cookie := auth.CreateCookie(id)

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
		Status:   model.OrderStatusRegistered,
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
	// err = h.accrualService.
	w.WriteHeader(http.StatusAccepted)
}

func (h *Handlers) GetOrdersInfo(w http.ResponseWriter, r *http.Request)      {}
func (h *Handlers) GetWithdrawalsInfo(w http.ResponseWriter, r *http.Request) {}
func (h *Handlers) GetUserBalance(w http.ResponseWriter, r *http.Request)     {}
func (h *Handlers) CreateWithdraw(w http.ResponseWriter, r *http.Request)     {}
