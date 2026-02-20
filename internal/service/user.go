package service

import (
	"context"
	"ecommerce-api-v2/internal/models"
	"errors"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jackc/pgx/v5/pgxpool"
)

type UserService struct {
	DB *pgxpool.Pool
}

func (s *UserService) CreateUser(ctx context.Context, req models.User) error {
	query := `INSERT INTO users (id, email, password_hash, role) VALUES ($1, $2, $3, $4)`

	_, err := s.DB.Exec(ctx, query, req.ID, req.Email, req.PasswordHash, req.Role)

	if err != nil {
		var pgErr *pgconn.PgError
		if errors.As(err, &pgErr) && pgErr.Code == "23505" {
			return errors.New("Email already exists")
		}
		return err
	}

	return nil
}

func (s *UserService) GetUserByEmail(ctx context.Context, email string) (*models.User, error) {
	query := `SELECT id, email, password_hash, role FROM users WHERE email = $1`

	var user models.User
	err := s.DB.QueryRow(ctx, query, email).Scan(&user.ID, &user.Email, &user.PasswordHash, &user.Role)

	if err != nil {
		if err == pgx.ErrNoRows {
			return nil, err
		}
		return nil, err
	}

	return &user, nil
}
