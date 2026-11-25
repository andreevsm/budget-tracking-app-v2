package service

import (
	"budget/internal/domain"
	"context"
	"fmt"

	"github.com/google/uuid"
)

func (s *Service) AccountStatistics(ctx context.Context, userID, accountID uuid.UUID, from, to domain.Date, bucket domain.StatBucket) (domain.PeriodStatistics, error) {
	if !from.Valid() || !to.Valid() {
		return domain.PeriodStatistics{}, fmt.Errorf("account statistics: %w: %v", domain.ErrInvalidInput, "from/to must be dates in YYYY-MM-DD")
	}

	if from > to {
		return domain.PeriodStatistics{}, fmt.Errorf("account statistics: %w: %v", domain.ErrInvalidInput, "from must not be after to")
	}

	if !bucket.Valid() {
		return domain.PeriodStatistics{}, fmt.Errorf("account statistics: %w: %v", domain.ErrInvalidInput, "bucket must be day|week|month|year")
	}

	accountExists, err := s.repo.AccountBelongsToUser(ctx, accountID, userID)
	if err != nil {
		return domain.PeriodStatistics{}, err
	}

	if !accountExists {
		return domain.PeriodStatistics{}, fmt.Errorf("account statistics: %w: %v", domain.ErrAccountNotFound, "account doesn't exist")
	}

	return s.repo.AccountStatistics(ctx, accountID, userID, from, to, bucket)
}
