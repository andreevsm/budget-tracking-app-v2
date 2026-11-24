package service

import (
	"budget/internal/domain"
	"context"
	"fmt"

	"github.com/google/uuid"
)

func (s *Service) CreateTransfer(ctx context.Context, userID uuid.UUID, in domain.CreateTransferInput) (uuid.UUID, error) {
	if in.Amount.Sign() != 1 {
		return uuid.UUID{}, fmt.Errorf("create transfer: %w: %v", domain.ErrInvalidInput, "amount must be > 0")
	}

	if in.FromAccountID == uuid.Nil {
		return uuid.UUID{}, fmt.Errorf("create transfer: %w: %v", domain.ErrInvalidInput, "from_account_id must not be empty")
	}

	if in.ToAccountID == uuid.Nil {
		return uuid.UUID{}, fmt.Errorf("create transfer: %w: %v", domain.ErrInvalidInput, "to_account_id must not be empty")
	}

	if in.FromAccountID == in.ToAccountID {
		return uuid.UUID{}, fmt.Errorf("create transfer: %w: %v", domain.ErrInvalidInput, "from_account_id and to_account_id must be different")
	}

	if !in.OccurredOn.Valid() {
		return uuid.UUID{}, fmt.Errorf("create transfer: %w: %v", domain.ErrInvalidInput, "occurred_on must be date")
	}

	fromAccountExists, err := s.repo.AccountBelongsToUser(ctx, in.FromAccountID, userID)

	if err != nil {
		return uuid.UUID{}, err
	}

	if !fromAccountExists {
		return uuid.UUID{}, fmt.Errorf("create transfer: %w: %v", domain.ErrAccountNotFound, "account doesn't exist by from_account_id")
	}

	toAccountExists, err := s.repo.AccountBelongsToUser(ctx, in.ToAccountID, userID)

	if err != nil {
		return uuid.UUID{}, err
	}

	if !toAccountExists {
		return uuid.UUID{}, fmt.Errorf("create transfer: %w: %v", domain.ErrAccountNotFound, "account doesn't exist by to_account_id")
	}

	transferGroupID := uuid.New()

	_, err = s.repo.CreateTransfer(ctx, transferGroupID, userID, in)

	if err != nil {
		return uuid.UUID{}, err
	}

	return transferGroupID, nil
}
