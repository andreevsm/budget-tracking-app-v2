package repository

import (
	"context"
	"errors"
	"fmt"

	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jackc/pgx/v5/pgxpool"

	"budget/internal/domain"
)

type Repository struct {
	pool *pgxpool.Pool
}

func NewRepository(pool *pgxpool.Pool) *Repository {
	return &Repository{pool: pool}
}

func (r *Repository) ListCurrencies(ctx context.Context) ([]domain.Currency, error) {
	const query = `SELECT code, name, symbol, minor_units FROM currencies ORDER BY code`

	rows, err := r.pool.Query(ctx, query)

	if err != nil {
		return nil, fmt.Errorf("query currencies: %w", err)
	}

	defer rows.Close()

	currencies := make([]domain.Currency, 0)

	for rows.Next() {
		var currency domain.Currency

		if err := rows.Scan(&currency.Code, &currency.Name, &currency.Symbol, &currency.MinorUnits); err != nil {
			return nil, fmt.Errorf("scan currency: %w", err)
		}

		currencies = append(currencies, currency)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("iterate currencies: %w", err)
	}

	return currencies, nil
}

func (r *Repository) CreateAccount(ctx context.Context, in domain.CreateAccountInput) (domain.Account, error) {
	const query = `INSERT INTO accounts (user_id, name, account_type, currency) VALUES ($1, $2, $3, $4) RETURNING id, created_at`

	account := domain.Account{
		UserID:      in.UserID,
		Name:        in.Name,
		AccountType: in.AccountType,
		Currency:    in.Currency,
	}

	err := r.pool.QueryRow(ctx, query, in.UserID, in.Name, in.AccountType, in.Currency).Scan(&account.ID, &account.CreatedAt)

	if err == nil {
		return account, nil
	}

	if pgErr, ok := errors.AsType[*pgconn.PgError](err); ok {
		switch pgErr.Code {
		case "23505":
			return domain.Account{}, fmt.Errorf("insert account: %w: %w", domain.ErrDuplicateName, err)
		case "23503":
			if pgErr.ConstraintName == "accounts_user_id_fkey" {
				return domain.Account{}, fmt.Errorf("insert account: %w: %w", domain.ErrUserNotFound, err)
			}

			if pgErr.ConstraintName == "accounts_currency_fkey" {
				return domain.Account{}, fmt.Errorf("insert account: %w: %w", domain.ErrInvalidReference, err)
			}
		case "22P02":
			return domain.Account{}, fmt.Errorf("insert account: %w: %w", domain.ErrInvalidInput, err)
		}
	}

	return domain.Account{}, fmt.Errorf("insert account: %w", err)
}
