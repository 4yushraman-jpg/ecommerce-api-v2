package handlers

import (
	"bytes"
	"context"
	"ecommerce-api-v2/internal/middleware"
	"ecommerce-api-v2/internal/models"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/google/uuid"
)

func TestAddToCart_UpsertLogic(t *testing.T) {
	db := setupTestDB()
	defer db.Close()
	handler := &CartHandler{DB: db}

	userID := uuid.New()
	productID := uuid.New()

	_, err := db.Exec(context.Background(), `
		INSERT INTO users (id, email, password_hash, role) 
		VALUES ($1, 'cartuser@example.com', 'hash', 'customer')
	`, userID)
	if err != nil {
		t.Fatalf("Failed to insert test user: %v", err)
	}

	_, err = db.Exec(context.Background(), `
		INSERT INTO products (id, name, price, stock_quantity) 
		VALUES ($1, 'Test Product', 1000, 50)
	`, productID)
	if err != nil {
		t.Fatalf("Failed to insert test product: %v", err)
	}

	addToCart := func(quantity int) *httptest.ResponseRecorder {
		reqBody := models.AddToCartRequest{
			ProductID: productID.String(),
			Quantity:  quantity,
		}
		bodyBytes, _ := json.Marshal(reqBody)

		req := httptest.NewRequest(http.MethodPost, "/api/v1/cart", bytes.NewReader(bodyBytes))

		claims := middleware.UserClaims{
			UserID: userID.String(),
			Role:   "customer",
		}
		ctx := context.WithValue(req.Context(), middleware.UserContextKey, claims)
		req = req.WithContext(ctx)

		w := httptest.NewRecorder()
		handler.AddToCartHandler(w, req)
		return w
	}

	w1 := addToCart(2)
	if w1.Code != http.StatusOK {
		t.Fatalf("Expected 200 OK for first add, got %d", w1.Code)
	}

	w2 := addToCart(3)
	if w2.Code != http.StatusOK {
		t.Fatalf("Expected 200 OK for second add, got %d", w2.Code)
	}

	var totalQuantity int
	err = db.QueryRow(
		context.Background(),
		"SELECT quantity FROM cart_items WHERE user_id = $1 AND product_id = $2",
		userID, productID,
	).Scan(&totalQuantity)

	if err != nil {
		t.Fatalf("Failed to query cart_items: %v", err)
	}

	if totalQuantity != 5 {
		t.Errorf("UPSERT failed! Expected quantity to be 5, but got %d", totalQuantity)
	}
}
