# Ecommerce API v2

A RESTful ecommerce API built with Go, Chi router, and PostgreSQL. Supports user authentication, product management, shopping cart, and order processing with role-based access control (customer and admin).

## Features

- **Authentication** — JWT-based auth with user registration and login
- **Products** — Public product listing; admin-only create, update, delete
- **Cart** — Add items, view cart, remove items (requires auth)
- **Orders** — Checkout, order history, and admin order status updates
- **Roles** — Customer and admin roles with different permissions
- **PostgreSQL** — Database with migrations

## Prerequisites

- Go 1.25+
- PostgreSQL
- `DATABASE_URL` and `JWT_SECRET` environment variables

## Setup

1. **Clone and install dependencies**

   ```bash
   go mod download
   ```

2. **Configure environment**

   Create a `.env` file in the project root:

   ```
   DATABASE_URL=postgres://user:password@localhost:5432/ecommerce?sslmode=disable
   JWT_SECRET=your-secret-key
   ```

3. **Run database migrations**

   Execute the SQL in `internal/database/migrations/001_init.sql` against your PostgreSQL database to create tables.

4. **Start the server**

   ```bash
   go run ./cmd/api
   ```

   The API listens on `http://localhost:8080`.

## API Endpoints

Base URL: `/api/v1`

| Method | Endpoint | Auth | Description |
|--------|----------|------|-------------|
| POST | `/users/register` | No | Register a new user |
| POST | `/users/login` | No | Login and receive JWT |
| GET | `/products` | No | List all products |
| GET | `/products/{id}` | No | Get a product by ID |
| POST | `/cart` | Yes | Add item to cart |
| GET | `/cart` | Yes | Get current cart |
| DELETE | `/cart/{product_id}` | Yes | Remove item from cart |
| POST | `/checkout` | Yes | Create order from cart |
| GET | `/orders` | Yes | Get order history |
| POST | `/products` | Admin | Create product |
| PUT | `/products/{id}` | Admin | Update product |
| DELETE | `/products/{id}` | Admin | Delete product |
| PUT | `/orders/{id}/status` | Admin | Update order status |

### Authentication

Protected routes require the `Authorization: Bearer <token>` header with a valid JWT. Use the token from `/users/login`.

### Admin routes

Admin-only routes require a user with `role: "admin"`.

## Project structure

```
├── cmd/
│   └── api/
│       └── main.go          # Application entry point
├── internal/
│   ├── database/
│   │   ├── db.go            # Database connection
│   │   └── migrations/      # SQL migrations
│   ├── handlers/            # HTTP handlers
│   ├── middleware/          # Auth and admin middleware
│   └── models/              # Data models
├── go.mod
└── go.sum
```

## Tech stack

- **Router** — [Chi](https://github.com/go-chi/chi)
- **Database** — [pgx](https://github.com/jackc/pgx/v5)
- **JWT** — [golang-jwt/jwt](https://github.com/golang-jwt/jwt)
- **Env** — [godotenv](https://github.com/joho/godotenv)
- **Project URL** — [project URL](https://roadmap.sh/projects/ecommerce-api)
