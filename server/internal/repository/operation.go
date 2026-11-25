package repository

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jackc/pgx/v5/pgtype"

	"budget/internal/domain"
)

func (r *Repository) CreateOperation(ctx context.Context, userID uuid.UUID, in domain.CreateOperationInput) (domain.Operation, error) {
	const query = `INSERT INTO transactions (user_id, account_id, category_id, type, amount, occurred_on, payee, description) VALUES ($1, $2, $3, $4, $5, $6, $7, $8) RETURNING id, created_at`

	operation := domain.Operation{
		UserID:      userID,
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
	const query = `SELECT id, user_id, account_id, category_id, type, amount, occurred_on, payee, description, transfer_group_id, created_at FROM transactions WHERE account_id = $1 AND deleted_at IS NULL ORDER BY occurred_on DESC`

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

		if err := rows.Scan(&operation.ID, &operation.UserID, &operation.AccountID, &operation.CategoryID, &operation.Type, &amount, &occurredOn, &operation.Payee, &operation.Description, &operation.TransferGroupID, &operation.CreatedAt); err != nil {
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

func (r *Repository) GetOperation(ctx context.Context, operationID, userID uuid.UUID) (domain.Operation, error) {
	const query = `SELECT id, user_id, account_id, category_id, type, amount, occurred_on, payee, description, transfer_group_id, created_at FROM transactions WHERE id = $1 AND user_id = $2 AND deleted_at IS NULL`

	var operation domain.Operation
	var amount pgtype.Numeric
	var occurredOn time.Time
	var transferGroupID *uuid.UUID

	err := r.pool.QueryRow(ctx, query, operationID, userID).Scan(&operation.ID, &operation.UserID, &operation.AccountID, &operation.CategoryID, &operation.Type, &amount, &occurredOn, &operation.Payee, &operation.Description, &transferGroupID, &operation.CreatedAt)

	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return domain.Operation{}, fmt.Errorf("get operation: %w", domain.ErrOperationNotFound)
		}

		return domain.Operation{}, fmt.Errorf("get operation: %w", err)
	}

	money, err := amount.Value()
	if err != nil {
		return domain.Operation{}, fmt.Errorf("numeric to string: %w", err)
	}

	operation.Amount = domain.Money(money.(string))
	operation.OccurredOn = domain.Date(occurredOn.Format("2006-01-02"))
	operation.TransferGroupID = transferGroupID
	return operation, nil
}

func (r *Repository) UpdateOperation(ctx context.Context, operationID, userID uuid.UUID, in domain.UpdateOperationInput) error {
	const query = `UPDATE transactions SET amount = $1, category_id = $2, occurred_on = $3, payee = $4, description = $5 WHERE id = $6 AND user_id = $7 AND deleted_at IS NULL`

	tag, err := r.pool.Exec(ctx, query, in.Amount, in.CategoryID, in.OccurredOn, in.Payee, in.Description, operationID, userID)

	if err != nil {
		return fmt.Errorf("update operation: %w", err)
	}

	if tag.RowsAffected() == 0 {
		return fmt.Errorf("update operation: %w", domain.ErrOperationNotFound)
	}

	return nil
}

func (r *Repository) UpdateTransfer(ctx context.Context, groupID, userID uuid.UUID, in domain.UpdateOperationInput) error {
	tx, err := r.pool.Begin(ctx)
	if err != nil {
		return err
	}
	defer tx.Rollback(ctx)

	const queryFrom = `UPDATE transactions SET amount = $1::numeric * -1, occurred_on = $2, payee = $3, description = $4 WHERE transfer_group_id = $5 AND user_id = $6 AND amount < 0 AND deleted_at IS NULL`

	outTag, err := tx.Exec(ctx, queryFrom, in.Amount, in.OccurredOn, in.Payee, in.Description, groupID, userID)

	if err != nil {
		return fmt.Errorf("update transfer out: %w", err)
	}

	const queryTo = `UPDATE transactions SET amount = $1, occurred_on = $2, payee = $3, description = $4 WHERE transfer_group_id = $5 AND user_id = $6 AND amount > 0 AND deleted_at IS NULL`
	inTag, err := tx.Exec(ctx, queryTo, in.Amount, in.OccurredOn, in.Payee, in.Description, groupID, userID)

	if err != nil {
		return fmt.Errorf("update transfer in: %w", err)
	}

	if outTag.RowsAffected() != 1 || inTag.RowsAffected() != 1 {
		return fmt.Errorf("update transfer: %w", domain.ErrOperationNotFound)
	}

	if err := tx.Commit(ctx); err != nil {
		return err
	}

	return nil
}

func (r *Repository) DeleteTransfer(ctx context.Context, groupID, userID uuid.UUID) error {
	const query = `UPDATE transactions SET deleted_at = now() WHERE transfer_group_id = $1 AND user_id = $2 AND deleted_at IS NULL`

	tag, err := r.pool.Exec(ctx, query, groupID, userID)
	if err != nil {
		return fmt.Errorf("soft delete transfer: %w", err)
	}

	if tag.RowsAffected() == 0 {
		return fmt.Errorf("soft delete transfer: %w", domain.ErrOperationNotFound)
	}

	return nil
}

func (r *Repository) DeleteOperation(ctx context.Context, operationID, userID uuid.UUID) error {
	const query = `UPDATE transactions SET deleted_at = now() WHERE id = $1 AND user_id = $2 AND deleted_at IS NULL`

	tag, err := r.pool.Exec(ctx, query, operationID, userID)
	if err != nil {
		return fmt.Errorf("soft delete operation: %w", err)
	}

	if tag.RowsAffected() == 0 {
		return fmt.Errorf("soft delete operation: %w", domain.ErrOperationNotFound)
	}

	return nil
}
