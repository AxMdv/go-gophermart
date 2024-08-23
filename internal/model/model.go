package model

import "time"

type User struct {
	UUID     string
	Login    string `json:"login"`
	Password string `json:"password"`
	Balance  Balance
}

type Order struct {
	ID         int
	UserUUID   string
	Status     string
	Accrual    int32
	UploadedAt time.Time
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
