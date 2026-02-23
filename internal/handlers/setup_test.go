package handlers

import (
	"context"
	"log"

	"github.com/jackc/pgx/v5/pgxpool"
)

func setupTestDB() *pgxpool.Pool {
	dsn := "postgres://user:password@localhost:5432/ecommerce_test"

	pool, err := pgxpool.New(context.Background(), dsn)
	if err != nil {
		log.Fatalf("Unable to connect to test database: %v", err)
	}

	_, err = pool.Exec(context.Background(), `
		TRUNCATE TABLE cart_items, order_items, orders, products, users CASCADE;
	`)
	if err != nil {
		log.Fatalf("Failed to truncate tables: %v", err)
	}

	return pool
}
