package handlers

import (
	"ecommerce-api-v2/internal/middleware"
	"ecommerce-api-v2/internal/models"
	"encoding/json"
	"net/http"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"
)

type OrderHandler struct {
	DB *pgxpool.Pool
}

func (h *OrderHandler) CheckoutHandler(w http.ResponseWriter, r *http.Request) {
	ctxValue := r.Context().Value(middleware.UserContextKey)
	claims, ok := ctxValue.(middleware.UserClaims)
	if !ok {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}
	userID, _ := uuid.Parse(claims.UserID)

	tx, err := h.DB.Begin(r.Context())
	if err != nil {
		http.Error(w, "Failed to start checkout process", http.StatusInternalServerError)
		return
	}
	defer tx.Rollback(r.Context())

	queryCart := `
		SELECT c.product_id, c.quantity, p.price, p.stock_quantity
		FROM cart_items c
		JOIN products p ON c.product_id = p.id
		WHERE c.user_id = $1
		FOR UPDATE OF p; 
	`
	rows, err := tx.Query(r.Context(), queryCart, userID)
	if err != nil {
		http.Error(w, "Error reading cart", http.StatusInternalServerError)
		return
	}

	type checkoutItem struct {
		ProductID uuid.UUID
		Quantity  int
		Price     int
		Stock     int
	}
	var items []checkoutItem
	totalAmount := 0

	for rows.Next() {
		var item checkoutItem
		if err := rows.Scan(&item.ProductID, &item.Quantity, &item.Price, &item.Stock); err != nil {
			rows.Close()
			http.Error(w, "Error parsing cart items", http.StatusInternalServerError)
			return
		}

		if item.Quantity > item.Stock {
			rows.Close()
			http.Error(w, "Insufficient stock for one or more items", http.StatusConflict)
			return
		}

		items = append(items, item)
		totalAmount += (item.Price * item.Quantity)
	}
	rows.Close()

	if len(items) == 0 {
		http.Error(w, "Your cart is empty", http.StatusBadRequest)
		return
	}

	var orderID uuid.UUID
	createOrderQuery := `
		INSERT INTO orders (user_id, total_amount, status)
		VALUES ($1, $2, 'pending')
		RETURNING id
	`
	err = tx.QueryRow(r.Context(), createOrderQuery, userID, totalAmount).Scan(&orderID)
	if err != nil {
		http.Error(w, "Failed to create order", http.StatusInternalServerError)
		return
	}

	insertOrderItemQuery := `
		INSERT INTO order_items (order_id, product_id, quantity, price_at_purchase)
		VALUES ($1, $2, $3, $4)
	`
	updateStockQuery := `
		UPDATE products SET stock_quantity = stock_quantity - $1 WHERE id = $2
	`

	for _, item := range items {
		if _, err := tx.Exec(r.Context(), insertOrderItemQuery, orderID, item.ProductID, item.Quantity, item.Price); err != nil {
			http.Error(w, "Failed to save order details", http.StatusInternalServerError)
			return
		}
		if _, err := tx.Exec(r.Context(), updateStockQuery, item.Quantity, item.ProductID); err != nil {
			http.Error(w, "Failed to update inventory", http.StatusInternalServerError)
			return
		}
	}

	clearCartQuery := `DELETE FROM cart_items WHERE user_id = $1`
	if _, err := tx.Exec(r.Context(), clearCartQuery, userID); err != nil {
		http.Error(w, "Failed to clear cart", http.StatusInternalServerError)
		return
	}

	if err := tx.Commit(r.Context()); err != nil {
		http.Error(w, "Failed to finalize checkout", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(models.CheckoutResponse{
		OrderID:     orderID.String(),
		TotalAmount: totalAmount,
		Status:      "pending",
		Message:     "Checkout successful! Your order has been placed.",
	})
}

func (h *OrderHandler) GetOrderHistoryHandler(w http.ResponseWriter, r *http.Request) {
	ctxValue := r.Context().Value(middleware.UserContextKey)
	claims, ok := ctxValue.(middleware.UserClaims)
	if !ok {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	query := `
		SELECT 
			o.id, o.total_amount, o.status, o.created_at,
			oi.product_id, p.name, oi.quantity, oi.price_at_purchase
		FROM orders o
		JOIN order_items oi ON o.id = oi.order_id
		JOIN products p ON oi.product_id = p.id
		WHERE o.user_id = $1
		ORDER BY o.created_at DESC, o.id
	`

	rows, err := h.DB.Query(r.Context(), query, claims.UserID)
	if err != nil {
		http.Error(w, "Database error", http.StatusInternalServerError)
		return
	}
	defer rows.Close()

	history := make([]models.OrderHistoryResponse, 0)

	for rows.Next() {
		var orderID, status, productID, productName string
		var totalAmount, quantity, priceAtPurchase int
		var createdAt time.Time

		if err := rows.Scan(
			&orderID, &totalAmount, &status, &createdAt,
			&productID, &productName, &quantity, &priceAtPurchase,
		); err != nil {
			http.Error(w, "Error reading order history", http.StatusInternalServerError)
			return
		}

		if len(history) == 0 || history[len(history)-1].OrderID != orderID {
			newOrder := models.OrderHistoryResponse{
				OrderID:     orderID,
				TotalAmount: totalAmount,
				Status:      status,
				CreatedAt:   createdAt,
				Items:       make([]models.OrderHistoryItemResponse, 0),
			}
			history = append(history, newOrder)
		}

		item := models.OrderHistoryItemResponse{
			ProductID:       productID,
			ProductName:     productName,
			Quantity:        quantity,
			PriceAtPurchase: priceAtPurchase,
		}

		lastIndex := len(history) - 1
		history[lastIndex].Items = append(history[lastIndex].Items, item)
	}

	if rows.Err() != nil {
		http.Error(w, "Error iterating over orders", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(history)
}

var validStatuses = map[string]bool{
	"pending":    true,
	"processing": true,
	"shipped":    true,
	"delivered":  true,
	"cancelled":  true,
}

func (h *OrderHandler) UpdateOrderStatusHandler(w http.ResponseWriter, r *http.Request) {
	orderIDStr := chi.URLParam(r, "id")
	orderID, err := uuid.Parse(orderIDStr)
	if err != nil {
		http.Error(w, "Invalid order ID format", http.StatusBadRequest)
		return
	}

	var req models.UpdateOrderStatusRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid JSON payload", http.StatusBadRequest)
		return
	}

	if !validStatuses[req.Status] {
		http.Error(w, "Invalid status. Allowed values: pending, processing, shipped, delivered, cancelled", http.StatusBadRequest)
		return
	}

	query := `UPDATE orders SET status = $1 WHERE id = $2`

	cmdTag, err := h.DB.Exec(r.Context(), query, req.Status, orderID)
	if err != nil {
		http.Error(w, "Database error while updating order", http.StatusInternalServerError)
		return
	}

	if cmdTag.RowsAffected() == 0 {
		http.Error(w, "Order not found", http.StatusNotFound)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(map[string]string{
		"message": "Order status updated to " + req.Status,
	})
}
