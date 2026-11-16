package repository

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jackc/pgx/v5/pgtype"

	"budget/internal/domain"
)

func (r *Repository) CreateOperation(ctx context.Context, in domain.CreateOperationInput) (domain.Operation, error) {
	const query = `INSERT INTO transactions (user_id, account_id, category_id, type, amount, occurred_on, payee, description) VALUES ($1, $2, $3, $4, $5, $6, $7, $8) RETURNING id, created_at`

	operation := domain.Operation{
		UserID:      in.UserID,
		AccountID:   in.AccountID,
		CategoryID:  in.CategoryID,
		Type:        in.Type,
		Amount:      in.Amount,
		OccurredOn:  in.OccurredOn,
		Payee:       in.Payee,
		Description: in.Description,
	}

	err := r.pool.QueryRow(ctx, query, operation.UserID, operation.AccountID, operation.CategoryID, operation.Type, operation.Amount, operation.OccurredOn, operation.Payee, operation.Description).Scan(&operation.ID, &operation.CreatedAt)

	if err == nil {
		return operation, nil
	}

	if pgErr, ok := errors.AsType[*pgconn.PgError](err); ok {
		switch pgErr.Code {
		case "23503":
			if pgErr.ConstraintName == "transactions_user_id_fkey" {
				return domain.Operation{}, fmt.Errorf("insert operation: %w: %v", domain.ErrUserNotFound, err)
			}
			if pgErr.ConstraintName == "transactions_account_id_fkey" {
				return domain.Operation{}, fmt.Errorf("insert operation: %w: %v", domain.ErrAccountNotFound, err)
			}
			if pgErr.ConstraintName == "transactions_category_id_fkey" {
				return domain.Operation{}, fmt.Errorf("insert operation: %w: %v", domain.ErrInvalidReference, err)
			}
		case "22P02":
			return domain.Operation{}, fmt.Errorf("insert operation: %w: %v", domain.ErrInvalidInput, err)
		}
	}

	return domain.Operation{}, fmt.Errorf("insert operation: %w", err)
}

func (r *Repository) ListOperationsByAccount(ctx context.Context, accountID uuid.UUID) ([]domain.Operation, error) {
	const query = `SELECT id, user_id, account_id, category_id, type, amount, occurred_on, payee, description, created_at FROM transactions WHERE account_id = $1 AND deleted_at IS NULL ORDER BY occurred_on DESC`

	rows, err := r.pool.Query(ctx, query, accountID)

	if err != nil {
		return nil, fmt.Errorf("query operations: %w", err)
	}

	defer rows.Close()

	operations := make([]domain.Operation, 0)

	for rows.Next() {
		var operation domain.Operation
		var amount pgtype.Numeric
		var occurredOn time.Time

		if err := rows.Scan(&operation.ID, &operation.UserID, &operation.AccountID, &operation.CategoryID, &operation.Type, &amount, &occurredOn, &operation.Payee, &operation.Description, &operation.CreatedAt); err != nil {
			return nil, fmt.Errorf("scan operation: %w", err)
		}

		money, err := amount.Value()

		if err != nil {
			return nil, fmt.Errorf("numeric to string: %w", err)
		}

		operation.Amount = domain.Money(money.(string))
		operation.OccurredOn = domain.Date(occurredOn.Format("2006-01-02"))
		operations = append(operations, operation)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("iterate operations: %w", err)
	}

	return operations, nil
}