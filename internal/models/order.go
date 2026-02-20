package models

import (
	"time"

	"github.com/google/uuid"
)

type Order struct {
	ID          uuid.UUID   `json:"id" db:"id"`
	UserID      uuid.UUID   `json:"user_id" db:"user_id"`
	TotalAmount int         `json:"total_amount" db:"total_amount"`
	Status      string      `json:"status" db:"status"`
	CreatedAt   time.Time   `json:"created_at" db:"created_at"`
	UpdatedAt   time.Time   `json:"updated_at" db:"updated_at"`
	Items       []OrderItem `json:"items,omitempty" db:"-"`
}

type OrderItem struct {
	ID              uuid.UUID `json:"id" db:"id"`
	OrderID         uuid.UUID `json:"order_id" db:"order_id"`
	ProductID       uuid.UUID `json:"product_id" db:"product_id"`
	Quantity        int       `json:"quantity" db:"quantity"`
	PriceAtPurchase int       `json:"price_at_purchase" db:"price_at_purchase"`
	CreatedAt       time.Time `json:"created_at" db:"created_at"`
}
