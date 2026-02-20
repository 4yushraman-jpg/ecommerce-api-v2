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

	r := chi.NewRouter()

	// // Essential Middlewares
	// r.Use(middleware.RequestID) // Injects a request ID into the context of each request
	// r.Use(middleware.RealIP)    // Sets a http.Request's RemoteAddr to either X-Forwarded-For or X-Real-IP
	// r.Use(middleware.Logger)    // Logs the start and end of each request with the elapsed time
	// r.Use(middleware.Recoverer) // Gracefully absorb panics and prints the stack trace

	// // Set a timeout value on the request context (ctx), that will signal through ctx.Done()
	// r.Use(middleware.Timeout(60 * time.Second))

	// 4. Define Routes
	// Grouping routes by version is a great practice for APIs
	r.Route("/api/v1", func(r chi.Router) {
		r.Post("/users/register", userHandler.RegisterUserHandler)
		r.Post("/users/login", userHandler.LoginUserHandler)
		r.Get("/products", productHandler.GetProductsHandler)
		r.Get("/products/{id}", productHandler.GetProductHandler)

		r.Group(func(r chi.Router) {
			r.Use(middleware.AuthMiddleware([]byte(jwtSecret)))

			r.Group(func(r chi.Router) {
				r.Use(middleware.AdminOnlyMiddleware)

				r.Post("/products", productHandler.CreateProductHandler)
				r.Put("/products/{id}", productHandler.UpdateProductHandler)
				r.Delete("/products/{id}", productHandler.DeleteProductHandler)
			})
		})
	})

	// 5. Configure the HTTP Server
	srv := &http.Server{
		Addr:    ":8080",
		Handler: r,
	}

	// 6. Start the server in a goroutine for Graceful Shutdown
	go func() {
		log.Printf("Starting server on port %s", srv.Addr)
		if err := srv.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
			log.Fatalf("Server failed to start: %v", err)
		}
	}()

	// 7. Wait for interrupt signal to gracefully shutdown the server
	quit := make(chan os.Signal, 1)
	// kill (no param) default send syscall.SIGTERM
	// kill -2 is syscall.SIGINT
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit
	log.Println("Shutting down server...")

	// The context is used to inform the server it has 5 seconds to finish
	// the request it is currently handling
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := srv.Shutdown(ctx); err != nil {
		log.Fatalf("Server forced to shutdown: %v", err)
	}

	log.Println("Server exiting gracefully")
}
