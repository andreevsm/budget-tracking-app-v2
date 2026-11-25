package service

import (
	"budget/internal/domain"
	"context"
	"fmt"

	"github.com/google/uuid"
)

func (s *Service) CreateOperation(ctx context.Context, userID uuid.UUID, in domain.CreateOperationInput) (domain.Operation, error) {
	if in.Amount.Valid() {
		if in.Type == domain.Expense && in.Amount.Sign() != -1 {
			return domain.Operation{}, fmt.Errorf("create operation: %w: %v", domain.ErrInvalidInput, "amount with type 'expense' must be < 0")
		}

		if in.Type == domain.Income && in.Amount.Sign() != 1 {
			return domain.Operation{}, fmt.Errorf("create operation: %w: %v", domain.ErrInvalidInput, "amount with type 'income' must be > 0")
		}
	} else {
		return domain.Operation{}, fmt.Errorf("create operation: %w: %v", domain.ErrInvalidInput, "amount must be correct")
	}

	if !in.OccurredOn.Valid() {
		return domain.Operation{}, fmt.Errorf("create operation: %w: %v", domain.ErrInvalidInput, "occurred_on must be correct")
	}

	if in.AccountID == uuid.Nil {
		return domain.Operation{}, fmt.Errorf("create operation: %w: %v", domain.ErrInvalidInput, "account_id is missing")
	}

	accountExists, err := s.repo.AccountBelongsToUser(ctx, in.AccountID, userID)

	if err != nil {
		return domain.Operation{}, err
	}

	if !accountExists {
		return domain.Operation{}, fmt.Errorf("create operation: %w: %v", domain.ErrAccountNotFound, "account doesn't exist")
	}

	operation, err := s.repo.CreateOperation(ctx, userID, domain.CreateOperationInput{
		AccountID:   in.AccountID,
		CategoryID:  in.CategoryID,
		Type:        in.Type,
		Amount:      in.Amount,
		OccurredOn:  in.OccurredOn,
		Payee:       in.Payee,
		Description: in.Description,
	})

	if err != nil {
		return domain.Operation{}, err
	}

	return operation, nil
}

func (s *Service) ListOperationsByAccount(ctx context.Context, userID uuid.UUID, accountID string) ([]domain.Operation, error) {
	accountUUID, err := uuid.Parse(accountID)

	if err != nil {
		return []domain.Operation{}, fmt.Errorf("list operations: %w: %v", domain.ErrInvalidInput, "account_id is not correct")
	}

	if accountUUID == uuid.Nil {
		return []domain.Operation{}, fmt.Errorf("list operations: %w: %v", domain.ErrInvalidInput, "account_id is missing")
	}

	accountExists, err := s.repo.AccountBelongsToUser(ctx, accountUUID, userID)

	if err != nil {
		return []domain.Operation{}, err
	}

	if !accountExists {
		return []domain.Operation{}, fmt.Errorf("list operations: %w: %v", domain.ErrAccountNotFound, "account doesn't exist")
	}

	operations, err := s.repo.ListOperationsByAccount(ctx, accountUUID)

	if err != nil {
		return []domain.Operation{}, err
	}

	return operations, nil
}

func (s *Service) GetOperation(ctx context.Context, userID, operationID uuid.UUID) (domain.Operation, error) {
	return s.repo.GetOperation(ctx, operationID, userID)
}

func (s *Service) UpdateOperation(ctx context.Context, userID, operationID uuid.UUID, in domain.UpdateOperationInput) (domain.Operation, error) {
	existing, err := s.repo.GetOperation(ctx, operationID, userID)
	if err != nil {
		return domain.Operation{}, err
	}

	if !in.Amount.Valid() {
		return domain.Operation{}, fmt.Errorf("update operation: %w: %v", domain.ErrInvalidInput, "amount must be correct")
	}

	if existing.Type == domain.Income && in.Amount.Sign() != 1 {
		return domain.Operation{}, fmt.Errorf("update operation: %w: %v", domain.ErrInvalidInput, "amount with type 'income' must be > 0")
	}

	if existing.Type == domain.Expense && in.Amount.Sign() != -1 {
		return domain.Operation{}, fmt.Errorf("update operation: %w: %v", domain.ErrInvalidInput, "amount with type 'expense' must be < 0")
	}

	if !in.OccurredOn.Valid() {
		return domain.Operation{}, fmt.Errorf("update operation: %w: %v", domain.ErrInvalidInput, "occurred_on must be correct")
	}

	if existing.Type == domain.Transfer && existing.TransferGroupID != nil {
		if in.Amount.Sign() != 1 {
			return domain.Operation{}, fmt.Errorf("update operation: %w: %v", domain.ErrInvalidInput, "transfer amount must be > 0")
		}
		if err := s.repo.UpdateTransfer(ctx, *existing.TransferGroupID, userID, in); err != nil {
			return domain.Operation{}, err
		}
	} else {
		if err := s.repo.UpdateOperation(ctx, operationID, userID, in); err != nil {
			return domain.Operation{}, err
		}
	}

	return s.repo.GetOperation(ctx, operationID, userID)
}

func (s *Service) DeleteOperation(ctx context.Context, userID, operationID uuid.UUID) error {
	existing, err := s.repo.GetOperation(ctx, operationID, userID)
	if err != nil {
		return err
	}

	if existing.Type == domain.Transfer && existing.TransferGroupID != nil {
		return s.repo.DeleteTransfer(ctx, *existing.TransferGroupID, userID)
	}

	return s.repo.DeleteOperation(ctx, operationID, userID)
}
