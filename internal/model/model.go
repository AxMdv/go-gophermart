package model

import "time"

type User struct {
	UUID     string
	Login    string `json:"login"`
	Password string `json:"password"`
	Balance  Balance
}

type Order struct {
	ID         string
	UserUUID   string
	Status     string
	Accrual    int32
	UploadedAt string
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
