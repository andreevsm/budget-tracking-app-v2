package repository

import (
	"context"
	"errors"
	"fmt"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jackc/pgx/v5/pgtype"

	"budget/internal/domain"
)

func (r *Repository) CreateAccount(ctx context.Context, userID uuid.UUID, in domain.CreateAccountInput) (domain.Account, error) {
	const query = `INSERT INTO accounts (user_id, name, account_type, currency) VALUES ($1, $2, $3, $4) RETURNING id, created_at`

	account := domain.Account{
		UserID:      userID,
		Name:        in.Name,
		AccountType: in.AccountType,
		Currency:    in.Currency,
	}

	err := r.pool.QueryRow(ctx, query, account.UserID, account.Name, account.AccountType, account.Currency).Scan(&account.ID, &account.CreatedAt)

	if err == nil {
		return account, nil
	}

	if pgErr, ok := errors.AsType[*pgconn.PgError](err); ok {
		switch pgErr.Code {
		case "23505":
			return domain.Account{}, fmt.Errorf("insert account: %w: %v", domain.ErrDuplicateName, err)
		case "23503":
			if pgErr.ConstraintName == "accounts_user_id_fkey" {
				return domain.Account{}, fmt.Errorf("insert account: %w: %v", domain.ErrUserNotFound, err)
			}

			if pgErr.ConstraintName == "accounts_currency_fkey" {
				return domain.Account{}, fmt.Errorf("insert account: %w: %v", domain.ErrInvalidReference, err)
			}
		case "22P02":
			return domain.Account{}, fmt.Errorf("insert account: %w: %v", domain.ErrInvalidInput, err)
		}
	}

	return domain.Account{}, fmt.Errorf("insert account: %w", err)
}

func (r *Repository) ListAccounts(ctx context.Context, userID uuid.UUID) ([]domain.Account, error) {
	const query = `SELECT id, name, account_type, currency, user_id, opening_balance, created_at FROM accounts WHERE user_id = $1 AND NOT is_archived ORDER BY name`

	rows, err := r.pool.Query(ctx, query, userID)

	if err != nil {
		return nil, fmt.Errorf("query accounts: %w", err)
	}

	defer rows.Close()

	accounts := make([]domain.Account, 0)

	for rows.Next() {
		var account domain.Account
		var openingBalance pgtype.Numeric

		if err := rows.Scan(&account.ID, &account.Name, &account.AccountType, &account.Currency, &account.UserID, &openingBalance, &account.CreatedAt); err != nil {
			return nil, fmt.Errorf("scan account: %w", err)
		}

		money, err := openingBalance.Value()

		if err != nil {
			return nil, fmt.Errorf("numeric to string: %w", err)
		}

		account.OpeningBalance = domain.Money(money.(string))
		accounts = append(accounts, account)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("iterate accounts: %w", err)
	}

	return accounts, nil
}

func (r *Repository) UserExists(ctx context.Context, userID uuid.UUID) (bool, error) {
	const query = `SELECT EXISTS(SELECT 1 FROM users WHERE id = $1)`

	var exists bool

	err := r.pool.QueryRow(ctx, query, userID).Scan(&exists)

	if err != nil {
		return false, fmt.Errorf("user exists: %w", err)
	}

	return exists, nil
}

func (r *Repository) AccountBelongsToUser(ctx context.Context, accountID, userID uuid.UUID) (bool, error) {
	const query = `SELECT EXISTS(SELECT 1 FROM accounts WHERE id = $1 AND user_id = $2 AND NOT is_archived)`

	var exists bool

	err := r.pool.QueryRow(ctx, query, accountID, userID).Scan(&exists)

	if err != nil {
		return false, fmt.Errorf("account exists: %w", err)
	}

	return exists, nil
}
