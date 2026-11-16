package service

import "budget/internal/repository"

type Service struct {
	repo *repository.Repository
}

func NewService(r *repository.Repository) *Service {
	return &Service{repo: r}
}