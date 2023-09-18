package models

import "time"

type Transaction struct {
	ID        int       `json:"id"`
	OrderId   int       `json:"order_id"`
	OrderNo   string    `json:"order_no"`
	Token     string    `json:"token"`
	Amount    int64     `json:"amount"`
	Status    string    `json:"status"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
	// Add other transaction-related fields here
}

type OrderWithTransaction struct {
	Order
	Transaction Transaction `json:"transaction"`
}
