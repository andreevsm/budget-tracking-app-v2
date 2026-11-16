package service

import (
	"budget/internal/domain"
	"context"
	"fmt"
	"strings"
	"unicode/utf8"

	"github.com/google/uuid"
)

var CategoryKinds = []domain.OperationType{
	domain.Income,
	domain.Expense,
}

func (s *Service) CreateCategory(ctx context.Context, in domain.CreateCategoryInput) (domain.Category, error) {
	if in.UserID == uuid.Nil {
		return domain.Category{}, fmt.Errorf("create category: %w: %v", domain.ErrInvalidInput, "user_id is empty")
	}

	if in.ParentID != nil {
		return domain.Category{}, fmt.Errorf("create category: %w: %v", domain.ErrInvalidInput, "parent_id must be null")
	}

	name := strings.TrimSpace(in.Name)

	if name == "" {
		return domain.Category{}, fmt.Errorf("create category: %w: %v", domain.ErrInvalidInput, "name should not be empty")
	}

	if utf8.RuneCountInString(name) > 100 {
		return domain.Category{}, fmt.Errorf("create category: %w: %v", domain.ErrInvalidInput, "name length should be at most 100 characters")
	}

	isValidCategoryKind := false

	for _, categoryKind := range CategoryKinds {
		if categoryKind == in.Kind {
			isValidCategoryKind = true
			break
		}
	}

	if !isValidCategoryKind {
		return domain.Category{}, fmt.Errorf("create category: %w: %v", domain.ErrInvalidInput, "kind should have correct type")
	}

	category, err := s.repo.CreateCategory(ctx, domain.CreateCategoryInput{
		Name:     name,
		UserID:   in.UserID,
		Kind:     in.Kind,
		Color:    in.Color,
		Icon:     in.Icon,
		ParentID: in.ParentID,
	})

	if err != nil {
		return domain.Category{}, err
	}

	return category, nil
}

func (s *Service) ListCategories(ctx context.Context, userID string) ([]domain.Category, error) {
	userUUID, err := uuid.Parse(userID)

	if err != nil {
		return []domain.Category{}, fmt.Errorf("list categories: %w: %v", domain.ErrInvalidInput, "user_id is not correct")
	}

	userExists, err := s.repo.UserExists(ctx, userUUID)

	if err != nil {
		return []domain.Category{}, fmt.Errorf("list categories: %w", err)
	}

	if !userExists {
		return []domain.Category{}, fmt.Errorf("list categories: %w: %v", domain.ErrUserNotFound, "user doesn't exist")
	}

	categories, err := s.repo.ListCategories(ctx, userUUID)

	if err != nil {
		return []domain.Category{}, err
	}

	return categories, nil
}