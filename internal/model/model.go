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
	UserUUID  string
	Current   float64 `json:"current"`
	Withdrawn float64 `json:"withdrawn"`
}

type Withdrawal struct {
	UserUUID    string
	OrderID     string    `json:"order"`
	Amount      float64   `json:"sum"`
	ProcessedAt time.Time `json:"processed_at"`
}

var (
	OrderStatusRegistered = "REGISTERED"
	OrderStatusInvalid    = "INVALID"
	OrderStatusProcessing = "PROCESSING"
	OrderStatusProcessed  = "PROCESSED"
)
