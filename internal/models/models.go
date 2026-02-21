package models

import (
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
)

type User struct {
	ID           uuid.UUID `json:"id" db:"id"`
	Email        string    `json:"email" db:"email"`
	PasswordHash string    `json:"-" db:"password_hash"`
	Role         string    `json:"role" db:"role"`
	CreatedAt    time.Time `json:"created_at" db:"created_at"`
	UpdatedAt    time.Time `json:"updated_at" db:"updated_at"`
}

type Product struct {
	ID            uuid.UUID `json:"id" db:"id"`
	Name          string    `json:"name" db:"name"`
	Description   string    `json:"description" db:"description"`
	Price         int       `json:"price" db:"price"`
	StockQuantity int       `json:"stock_quantity" db:"stock_quantity"`
	CreatedAt     time.Time `json:"created_at" db:"created_at"`
	UpdatedAt     time.Time `json:"updated_at" db:"updated_at"`
}

type Order struct {
	ID         uuid.UUID   `json:"id" db:"id"`
	UserID     uuid.UUID   `json:"user_id" db:"user_id"`
	TotalPrice int         `json:"total_price" db:"total_price"`
	Status     string      `json:"status" db:"status"`
	CreatedAt  time.Time   `json:"created_at" db:"created_at"`
	UpdatedAt  time.Time   `json:"updated_at" db:"upadted_at"`
	Items      []OrderItem `json:"items,omitempty" db:"-"`
}

type OrderItem struct {
	ID              uuid.UUID `json:"id" db:"id"`
	OrderID         uuid.UUID `json:"order_id" db:"order_id"`
	ProductID       uuid.UUID `json:"product_id" db:"product_id"`
	Quantity        int       `json:"quantity" db:"quantity"`
	PriceAtPurchase int       `json:"price_at_purchase" db:"price_at_purchase"`
	CreatedAt       time.Time `json:"created_at" db:"created_at"`
}

type CartItem struct {
	ID        uuid.UUID `json:"id" db:"id"`
	UserID    uuid.UUID `json:"user_id" db:"user_id"`
	ProductID uuid.UUID `json:"product_id" db:"product_id"`
	Quantity  int       `json:"quantity" db:"quantity"`
	CreatedAt time.Time `json:"created_at" db:"created_at"`
	UpdatedAt time.Time `json:"updated_at" db:"updated_at"`
}

type RegisterUserRequest struct {
	Email    string `json:"email"`
	Password string `json:"password"`
}

type LoginUserRequest struct {
	Email    string `json:"email"`
	Password string `json:"password"`
}

type Claims struct {
	UserID uuid.UUID `json:"user_id"`
	Role   string    `json:"role"`
	jwt.RegisteredClaims
}

type GetProductResponse struct {
	ID            string `json:"id"`
	Name          string `json:"name"`
	Description   string `json:"description"`
	Price         int    `json:"price"`
	StockQuantity int    `josn:"stock_quantity"`
}

type CreateProductRequest struct {
	Name          string `json:"name"`
	Description   string `json:"description"`
	Price         int    `json:"price"`
	StockQuantity int    `json:"stock_quantity"`
}

type AddToCartRequest struct {
	ProductID string `json:"product_id"`
	Quantity  int    `json:"quantity"`
}

type CartItemResponse struct {
	CartItemID string `json:"cart_item_id"`
	ProductID  string `json:"product_id"`
	Name       string `json:"name"`
	Price      int    `json:"price"`
	Quantity   int    `json:"quantity"`
	Subtotal   int    `json:"subtotal"`
}

type CartResponse struct {
	Items      []CartItemResponse `json:"items"`
	TotalPrice int                `json:"total_price"`
}

type CheckoutResponse struct {
	OrderID     string `json:"order_id"`
	TotalAmount int    `json:"total_amount"`
	Status      string `json:"status"`
	Message     string `json:"message"`
}

type OrderHistoryItemResponse struct {
	ProductID       string `json:"product_id"`
	ProductName     string `json:"product_name"`
	Quantity        int    `json:"quantity"`
	PriceAtPurchase int    `json:"price_at_purchase"`
}

type OrderHistoryResponse struct {
	OrderID     string                     `json:"order_id"`
	TotalAmount int                        `json:"total_amount"`
	Status      string                     `json:"status"`
	CreatedAt   time.Time                  `json:"created_at"`
	Items       []OrderHistoryItemResponse `json:"items"`
}

type UpdateOrderStatusRequest struct {
	Status string `json:"status"`
}
