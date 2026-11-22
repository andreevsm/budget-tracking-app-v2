package service

import (
	"budget/internal/config"
	"budget/internal/repository"
)

type Service struct {
	repo *repository.Repository
	conf *config.Config
}

func NewService(r *repository.Repository, conf *config.Config) *Service {
	return &Service{repo: r, conf: conf}
}
