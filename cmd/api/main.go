package main

import (
	"context"
	"errors"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"ecommerce-api-v2/internal/database"
	"ecommerce-api-v2/internal/handlers"
	"ecommerce-api-v2/internal/middleware"

	"github.com/go-chi/chi/v5"
	"github.com/joho/godotenv"
)

func main() {
	if err := godotenv.Load(); err != nil {
		log.Println("No .env file found, relying on system environment variables")
	}

	log.Println("Connecting to the database...")
	dbPool, err := database.ConnectDB()
	if err != nil {
		log.Fatalf("Failed to connect to database: %v", err)
	}
	defer dbPool.Close()
	log.Println("Database connection established")

	jwtSecret := os.Getenv("JWT_SECRET")
	if jwtSecret == "" {
		log.Fatal("JWT_SECRET environment variable is not set")
	}

	userHandler := &handlers.UserHandler{
		DB:        dbPool,
		JWTSecret: []byte(jwtSecret),
	}

	productHandler := &handlers.ProductHandler{
		DB: dbPool,
	}

	cartHandler := &handlers.CartHandler{
		DB: dbPool,
	}

	orderHandler := &handlers.OrderHandler{
		DB: dbPool,
	}

	r := chi.NewRouter()

	r.Route("/api/v1", func(r chi.Router) {
		r.Post("/users/register", userHandler.RegisterUserHandler)
		r.Post("/users/login", userHandler.LoginUserHandler)
		r.Get("/products", productHandler.GetProductsHandler)
		r.Get("/products/{id}", productHandler.GetProductHandler)

		r.Group(func(r chi.Router) {
			r.Use(middleware.AuthMiddleware([]byte(jwtSecret)))

			r.Post("/cart", cartHandler.AddToCartHandler)
			r.Get("/cart", cartHandler.GetCartHandler)
			r.Delete("/cart/{product_id}", cartHandler.RemoveFromCartHandler)

			r.Post("/checkout", orderHandler.CheckoutHandler)
			r.Get("/orders", orderHandler.GetOrderHistoryHandler)

			r.Group(func(r chi.Router) {
				r.Use(middleware.AdminOnlyMiddleware)

				r.Post("/products", productHandler.CreateProductHandler)
				r.Put("/products/{id}", productHandler.UpdateProductHandler)
				r.Delete("/products/{id}", productHandler.DeleteProductHandler)

				r.Put("/orders/{id}/status", orderHandler.UpdateOrderStatusHandler)
			})
		})
	})

	srv := &http.Server{
		Addr:    ":8080",
		Handler: r,
	}

	go func() {
		log.Printf("Starting server on port %s", srv.Addr)
		if err := srv.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
			log.Fatalf("Server failed to start: %v", err)
		}
	}()

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit
	log.Println("Shutting down server...")

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := srv.Shutdown(ctx); err != nil {
		log.Fatalf("Server forced to shutdown: %v", err)
	}

	log.Println("Server exiting gracefully")
}
