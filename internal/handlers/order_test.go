package handlers

import (
	"context"
	"ecommerce-api-v2/internal/middleware"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/google/uuid"
)

func TestCheckoutHandler_StockRollback(t *testing.T) {
	db := setupTestDB()
	defer db.Close()
	handler := &OrderHandler{DB: db}

	userID := uuid.New()
	productID := uuid.New()

	db.Exec(context.Background(), `
		INSERT INTO users (id, email, password_hash, role) 
		VALUES ($1, 'buyer@example.com', 'hash', 'customer')
	`, userID)

	db.Exec(context.Background(), `
		INSERT INTO products (id, name, price, stock_quantity) 
		VALUES ($1, 'Limited Edition Poster', 2000, 5)
	`, productID)

	db.Exec(context.Background(), `
		INSERT INTO cart_items (user_id, product_id, quantity) 
		VALUES ($1, $2, 10)
	`, userID, productID)

	req := httptest.NewRequest(http.MethodPost, "/api/v1/checkout", nil)

	ctx := context.WithValue(req.Context(), middleware.UserContextKey, middleware.UserClaims{
		UserID: userID.String(),
		Role:   "customer",
	})
	req = req.WithContext(ctx)

	w := httptest.NewRecorder()
	handler.CheckoutHandler(w, req)

	if w.Code != http.StatusConflict {
		t.Errorf("Expected 409 Conflict for insufficient stock, got %d", w.Code)
	}

	var orderCount int
	db.QueryRow(context.Background(), "SELECT COUNT(*) FROM orders WHERE user_id = $1", userID).Scan(&orderCount)
	if orderCount != 0 {
		t.Errorf("CRITICAL FAILURE: Transaction didn't roll back! An order was created.")
	}

	var currentStock int
	db.QueryRow(context.Background(), "SELECT stock_quantity FROM products WHERE id = $1", productID).Scan(&currentStock)
	if currentStock != 5 {
		t.Errorf("CRITICAL FAILURE: Transaction didn't roll back! Stock was illegally reduced to %d", currentStock)
	}

	var cartCount int
	db.QueryRow(context.Background(), "SELECT COUNT(*) FROM cart_items WHERE user_id = $1", userID).Scan(&cartCount)
	if cartCount != 1 {
		t.Errorf("CRITICAL FAILURE: Transaction didn't roll back! Cart was emptied despite the error.")
	}
}
