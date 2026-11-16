package service

import (
	"budget/internal/domain"
	"context"
)

func (s *Service) ListCurrencies(ctx context.Context) ([]domain.Currency, error) {
	return s.repo.ListCurrencies(ctx)
}