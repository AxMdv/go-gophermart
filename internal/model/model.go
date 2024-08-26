package model

import "time"

type User struct {
	UUID     string
	Login    string `json:"login"`
	Password string `json:"password"`
	Balance  Balance
}

type Order struct {
	ID         string    `json:"number"`
	UserUUID   string    `json:"userID,omitempty"`
	Status     string    `json:"status"`
	Accrual    float64   `json:"accrual,omitempty"`
	UploadedAt time.Time `json:"uploaded_at"`
}

type Balance struct {
	Current   float32
	Withdrawn float32
}

type Withdrawn struct {
	OrderID     string
	Amount      float32
	ProcessedAt time.Time
}

var (
	OrderStatusRegistered = "REGISTERED"
	OrderStatusInvalid    = "INVALID"
	OrderStatusProcessing = "PROCESSING"
	OrderStatusProcessed  = "PROCESSED"
)
