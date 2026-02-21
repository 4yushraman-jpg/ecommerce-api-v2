package handlers

import (
	"ecommerce-api-v2/internal/middleware"
	"ecommerce-api-v2/internal/models"
	"encoding/json"
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"
)

type CartHandler struct {
	DB *pgxpool.Pool
}

func (h *CartHandler) AddToCartHandler(w http.ResponseWriter, r *http.Request) {
	ctxValue := r.Context().Value(middleware.UserContextKey)
	claims, ok := ctxValue.(middleware.UserClaims)
	if !ok {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	userID, err := uuid.Parse(claims.UserID)
	if err != nil {
		http.Error(w, "Invalid user ID in token", http.StatusUnauthorized)
		return
	}

	var req models.AddToCartRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid JSON payload", http.StatusBadRequest)
		return
	}

	productID, err := uuid.Parse(req.ProductID)
	if err != nil {
		http.Error(w, "Invalid product ID format", http.StatusBadRequest)
		return
	}
	if req.Quantity <= 0 {
		http.Error(w, "Quantity must be greater than 0", http.StatusBadRequest)
		return
	}

	query := `
		INSERT INTO cart_items (user_id, product_id, quantity)
		VALUES ($1, $2, $3)
		ON CONFLICT (user_id, product_id) 
		DO UPDATE SET quantity = cart_items.quantity + EXCLUDED.quantity;
	`

	_, err = h.DB.Exec(r.Context(), query, userID, productID, req.Quantity)
	if err != nil {
		http.Error(w, "Could not add item to cart", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(map[string]string{
		"message": "Item added to cart successfully",
	})
}

func (h *CartHandler) GetCartHandler(w http.ResponseWriter, r *http.Request) {
	ctxValue := r.Context().Value(middleware.UserContextKey)
	claims, ok := ctxValue.(middleware.UserClaims)
	if !ok {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	query := `
		SELECT 
			ci.id, 
			ci.quantity, 
			p.id, 
			p.name, 
			p.price
		FROM cart_items ci
		JOIN products p ON ci.product_id = p.id
		WHERE ci.user_id = $1
		ORDER BY ci.created_at DESC
	`

	rows, err := h.DB.Query(r.Context(), query, claims.UserID)
	if err != nil {
		http.Error(w, "Database error", http.StatusInternalServerError)
		return
	}
	defer rows.Close()

	cart := models.CartResponse{
		Items:      make([]models.CartItemResponse, 0),
		TotalPrice: 0,
	}

	for rows.Next() {
		var item models.CartItemResponse

		if err := rows.Scan(&item.CartItemID, &item.Quantity, &item.ProductID, &item.Name, &item.Price); err != nil {
			http.Error(w, "Error reading cart items", http.StatusInternalServerError)
			return
		}

		item.Subtotal = item.Price * item.Quantity

		cart.TotalPrice += item.Subtotal

		cart.Items = append(cart.Items, item)
	}

	if rows.Err() != nil {
		http.Error(w, "Error iterating over cart", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(cart)
}

func (h *CartHandler) RemoveFromCartHandler(w http.ResponseWriter, r *http.Request) {
	ctxValue := r.Context().Value(middleware.UserContextKey)
	claims, ok := ctxValue.(middleware.UserClaims)
	if !ok {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	productIDStr := chi.URLParam(r, "product_id")
	productID, err := uuid.Parse(productIDStr)
	if err != nil {
		http.Error(w, "Invalid product ID format", http.StatusBadRequest)
		return
	}

	query := `DELETE FROM cart_items WHERE user_id = $1 AND product_id = $2`

	cmdTag, err := h.DB.Exec(r.Context(), query, claims.UserID, productID)
	if err != nil {
		http.Error(w, "Database error while removing item", http.StatusInternalServerError)
		return
	}

	if cmdTag.RowsAffected() == 0 {
		http.Error(w, "Item not found in your cart", http.StatusNotFound)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(map[string]string{
		"message": "Item removed from cart successfully",
	})
}
