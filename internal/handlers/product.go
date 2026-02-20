package handlers

import (
	"ecommerce-api-v2/internal/models"
	"encoding/json"
	"errors"
	"net/http"
	"strconv"
	"strings"

	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

type ProductHandler struct {
	DB *pgxpool.Pool
}

func (h *ProductHandler) CreateProductHandler(w http.ResponseWriter, r *http.Request) {
	var req models.CreateProductRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid JSON payload", http.StatusBadRequest)
		return
	}

	req.Name = strings.TrimSpace(req.Name)
	if req.Name == "" || req.Price <= 0 || req.StockQuantity < 0 {
		http.Error(w, "Invalid product details: name is required, price must be > 0, stock cannot be negative", http.StatusBadRequest)
		return
	}

	productID := uuid.New()

	query := `
		INSERT INTO products (id, name, description, price, stock_quantity)
		VALUES ($1, $2, $3, $4, $5)
	`

	_, err := h.DB.Exec(
		r.Context(),
		query,
		productID,
		req.Name,
		req.Description,
		req.Price,
		req.StockQuantity,
	)

	if err != nil {
		http.Error(w, "Could not create product", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(map[string]string{
		"message":    "Product created successfully",
		"product_id": productID.String(),
	})
}

func (h *ProductHandler) GetProductsHandler(w http.ResponseWriter, r *http.Request) {
	limit := 20
	offset := 0

	if l := r.URL.Query().Get("limit"); l != "" {
		if parsedLimit, err := strconv.Atoi(l); err == nil && parsedLimit > 0 && parsedLimit <= 100 {
			limit = parsedLimit
		}
	}

	if p := r.URL.Query().Get("page"); p != "" {
		if parsedPage, err := strconv.Atoi(p); err == nil && parsedPage > 1 {
			offset = (parsedPage - 1) * limit
		}
	}

	query := `
		SELECT id, name, description, price, stock_quantity 
		FROM products
		ORDER BY created_at DESC 
		LIMIT $1 OFFSET $2
	`

	rows, err := h.DB.Query(r.Context(), query, limit, offset)
	if err != nil {
		http.Error(w, "Database error", http.StatusInternalServerError)
		return
	}
	defer rows.Close()

	products := make([]models.GetProductResponse, 0)

	for rows.Next() {
		var p models.GetProductResponse
		if err := rows.Scan(&p.ID, &p.Name, &p.Description, &p.Price, &p.StockQuantity); err != nil {
			http.Error(w, "Error fetching rows", http.StatusInternalServerError)
			return
		}
		products = append(products, p)
	}

	if rows.Err() != nil {
		http.Error(w, "Error iterating over products", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(products)
}

func (h *ProductHandler) GetProductHandler(w http.ResponseWriter, r *http.Request) {
	productID := chi.URLParam(r, "id")
	if _, err := uuid.Parse(productID); err != nil {
		http.Error(w, "Invalid product ID format", http.StatusBadRequest)
		return
	}

	query := `SELECT id, name, description, price, stock_quantity FROM products WHERE id = $1`
	var p models.GetProductResponse

	err := h.DB.QueryRow(r.Context(), query, productID).Scan(
		&p.ID, &p.Name, &p.Description, &p.Price, &p.StockQuantity,
	)

	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			http.Error(w, "Product not found", http.StatusNotFound)
			return
		}
		http.Error(w, "Database error", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(p)
}

func (h *ProductHandler) UpdateProductHandler(w http.ResponseWriter, r *http.Request) {
	productID := chi.URLParam(r, "id")
	if _, err := uuid.Parse(productID); err != nil {
		http.Error(w, "Invalid product ID format", http.StatusBadRequest)
		return
	}

	var req models.CreateProductRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid JSON payload", http.StatusBadRequest)
		return
	}

	if req.Name == "" || req.Price <= 0 || req.StockQuantity < 0 {
		http.Error(w, "Invalid product details", http.StatusBadRequest)
		return
	}

	query := `
		UPDATE products 
		SET name = $1, description = $2, price = $3, stock_quantity = $4 
		WHERE id = $5
	`

	cmdTag, err := h.DB.Exec(r.Context(), query, req.Name, req.Description, req.Price, req.StockQuantity, productID)
	if err != nil {
		http.Error(w, "Could not update product", http.StatusInternalServerError)
		return
	}

	if cmdTag.RowsAffected() == 0 {
		http.Error(w, "Product not found", http.StatusNotFound)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(map[string]string{
		"message": "Product updated successfully",
	})
}

func (h *ProductHandler) DeleteProductHandler(w http.ResponseWriter, r *http.Request) {
	productID := chi.URLParam(r, "id")
	if _, err := uuid.Parse(productID); err != nil {
		http.Error(w, "Invalid product ID format", http.StatusBadRequest)
		return
	}

	query := `DELETE FROM products WHERE id = $1`

	cmdTag, err := h.DB.Exec(r.Context(), query, productID)
	if err != nil {
		http.Error(w, "Could not delete product. It may be part of an existing order.", http.StatusConflict)
		return
	}

	if cmdTag.RowsAffected() == 0 {
		http.Error(w, "Product not found", http.StatusNotFound)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(map[string]string{
		"message": "Product deleted successfully",
	})
}
