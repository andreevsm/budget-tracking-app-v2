package repository

import (
	"context"
	"errors"
	"fmt"

	"budget/internal/domain"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
)

func (r *Repository) GetUserByEmail(ctx context.Context, email string) (domain.User, string, error) {
	const query = `SELECT id, email, password_hash, display_name, base_currency, created_at FROM users WHERE email = $1`

	var user domain.User
	var passwordHash string

	err := r.pool.QueryRow(ctx, query, email).Scan(&user.ID, &user.Email, &passwordHash, &user.DisplayName, &user.BaseCurrency, &user.CreatedAt)

	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return domain.User{}, "", domain.ErrUserNotFound
		}

		return domain.User{}, "", fmt.Errorf("query user: %w", err)
	}

	return user, passwordHash, nil
}

func (r *Repository) CreateUser(ctx context.Context, in domain.RegisterInput, passwordHash string) (domain.User, error) {
	const query = `INSERT INTO users (email, password_hash, display_name, base_currency) VALUES ($1, $2, $3, $4) RETURNING id, created_at`

	user := domain.User{
		Email:        in.Email,
		DisplayName:  in.DisplayName,
		BaseCurrency: in.BaseCurrency,
	}

	err := r.pool.QueryRow(ctx, query, user.Email, passwordHash, user.DisplayName, user.BaseCurrency).Scan(&user.ID, &user.CreatedAt)

	if err == nil {
		return user, nil
	}

	if pgErr, ok := errors.AsType[*pgconn.PgError](err); ok {
		if pgErr.Code == "23505" {
			return domain.User{}, fmt.Errorf("insert user: %w: %v", domain.ErrEmailTaken, err)
		}
	}

	return domain.User{}, fmt.Errorf("insert user: %w", err)
}
