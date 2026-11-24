package repository

import (
	"context"
	"errors"
	"fmt"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgconn"

	"budget/internal/domain"
)

func (r *Repository) CreateTransfer(ctx context.Context, transferGroupID uuid.UUID, userID uuid.UUID, in domain.CreateTransferInput) (uuid.UUID, error) {
	tx, err := r.pool.Begin(ctx)

	if err != nil {
		return uuid.UUID{}, err
	}

	defer tx.Rollback(ctx)

	const queryFrom = `INSERT INTO transactions (user_id, account_id, category_id, type, amount, occurred_on, payee, description, transfer_group_id) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9)`

	_, err = tx.Exec(ctx, queryFrom, userID, in.FromAccountID, nil, "transfer", in.Amount.Negated(), in.OccurredOn, in.Payee, in.Description, transferGroupID)

	if err != nil {
		if pgErr, ok := errors.AsType[*pgconn.PgError](err); ok {
			switch pgErr.Code {
			case "23503":
				if pgErr.ConstraintName == "transactions_user_id_fkey" {
					return uuid.UUID{}, fmt.Errorf("insert transaction from: %w: %v", domain.ErrUserNotFound, err)
				}
				if pgErr.ConstraintName == "transactions_account_id_fkey" {
					return uuid.UUID{}, fmt.Errorf("insert transaction from: %w: %v", domain.ErrAccountNotFound, err)
				}

			case "22P02":
				return uuid.UUID{}, fmt.Errorf("insert transaction from: %w: %v", domain.ErrInvalidInput, err)
			}
		}

		return uuid.UUID{}, fmt.Errorf("insert transaction from: %w", err)
	}

	const queryTo = `INSERT INTO transactions (user_id, account_id, category_id, type, amount, occurred_on, payee, description, transfer_group_id) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9)`

	_, err = tx.Exec(ctx, queryTo, userID, in.ToAccountID, nil, "transfer", in.Amount, in.OccurredOn, in.Payee, in.Description, transferGroupID)

	if err != nil {

		if pgErr, ok := errors.AsType[*pgconn.PgError](err); ok {
			switch pgErr.Code {
			case "23503":
				if pgErr.ConstraintName == "transactions_user_id_fkey" {
					return uuid.UUID{}, fmt.Errorf("insert transaction to: %w: %v", domain.ErrUserNotFound, err)
				}
				if pgErr.ConstraintName == "transactions_account_id_fkey" {
					return uuid.UUID{}, fmt.Errorf("insert transaction to: %w: %v", domain.ErrAccountNotFound, err)
				}

			case "22P02":
				return uuid.UUID{}, fmt.Errorf("insert transaction to: %w: %v", domain.ErrInvalidInput, err)
			}
		}

		return uuid.UUID{}, fmt.Errorf("insert transaction to: %w", err)
	}

	if err := tx.Commit(ctx); err != nil {
		return uuid.UUID{}, err
	}

	return transferGroupID, nil
}
