package repository

import (
	"context"
	"errors"
	"fmt"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgconn"

	"budget/internal/domain"
)

func (r *Repository) CreateCategory(ctx context.Context, in domain.CreateCategoryInput) (domain.Category, error) {
	const query = `INSERT INTO categories (name, kind, color, icon, parent_id, user_id) VALUES ($1, $2, $3, $4, $5, $6) RETURNING id, created_at`

	category := domain.Category{
		Name:     in.Name,
		Kind:     in.Kind,
		Color:    in.Color,
		Icon:     in.Icon,
		ParentID: in.ParentID,
		UserID:   in.UserID,
	}

	err := r.pool.QueryRow(ctx, query, in.Name, in.Kind, in.Color, in.Icon, in.ParentID, in.UserID).Scan(&category.ID, &category.CreatedAt)

	if err == nil {
		return category, nil
	}

	if pgErr, ok := errors.AsType[*pgconn.PgError](err); ok {
		switch pgErr.Code {
		case "23505":
			return domain.Category{}, fmt.Errorf("insert category: %w: %v", domain.ErrDuplicateName, err)
		case "23503":
			if pgErr.ConstraintName == "categories_user_id_fkey" {
				return domain.Category{}, fmt.Errorf("insert category: %w: %v", domain.ErrUserNotFound, err)
			}
		case "22P02":
			return domain.Category{}, fmt.Errorf("insert category: %w: %v", domain.ErrInvalidInput, err)
		}
	}

	return domain.Category{}, fmt.Errorf("insert category: %w", err)
}

func (r *Repository) ListCategories(ctx context.Context, userID uuid.UUID) ([]domain.Category, error) {
	const query = `SELECT id, name, user_id, kind, color, icon, parent_id, created_at FROM categories WHERE user_id = $1 AND NOT is_archived ORDER BY kind, name`

	rows, err := r.pool.Query(ctx, query, userID)

	if err != nil {
		return nil, fmt.Errorf("query categories: %w", err)
	}

	defer rows.Close()

	categories := make([]domain.Category, 0)

	for rows.Next() {
		var category domain.Category

		if err := rows.Scan(&category.ID, &category.Name, &category.UserID, &category.Kind, &category.Color, &category.Icon, &category.ParentID, &category.CreatedAt); err != nil {
			return nil, fmt.Errorf("scan category: %w", err)
		}

		categories = append(categories, category)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("iterate categories: %w", err)
	}

	return categories, nil
}