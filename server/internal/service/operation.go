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
