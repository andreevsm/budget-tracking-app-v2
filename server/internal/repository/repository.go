package repository

import (
	"context"
	"errors"
	"fmt"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jackc/pgx/v5/pgtype"
	"github.com/jackc/pgx/v5/pgxpool"

	"budget/internal/domain"
)

type Repository struct {
	pool *pgxpool.Pool
}

func NewRepository(pool *pgxpool.Pool) *Repository {
	return &Repository{pool: pool}
}

func (r *Repository) ListCurrencies(ctx context.Context) ([]domain.Currency, error) {
	const query = `SELECT code, name, symbol, minor_units FROM currencies ORDER BY code`

	rows, err := r.pool.Query(ctx, query)

	if err != nil {
		return nil, fmt.Errorf("query currencies: %w", err)
	}

	defer rows.Close()

	currencies := make([]domain.Currency, 0)

	for rows.Next() {
		var currency domain.Currency

		if err := rows.Scan(&currency.Code, &currency.Name, &currency.Symbol, &currency.MinorUnits); err != nil {
			return nil, fmt.Errorf("scan currency: %w", err)
		}

		currencies = append(currencies, currency)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("iterate currencies: %w", err)
	}

	return currencies, nil
}

func (r *Repository) CreateAccount(ctx context.Context, in domain.CreateAccountInput) (domain.Account, error) {
	const query = `INSERT INTO accounts (user_id, name, account_type, currency) VALUES ($1, $2, $3, $4) RETURNING id, created_at`

	account := domain.Account{
		UserID:      in.UserID,
		Name:        in.Name,
		AccountType: in.AccountType,
		Currency:    in.Currency,
	}

	err := r.pool.QueryRow(ctx, query, in.UserID, in.Name, in.AccountType, in.Currency).Scan(&account.ID, &account.CreatedAt)

	if err == nil {
		return account, nil
	}

	if pgErr, ok := errors.AsType[*pgconn.PgError](err); ok {
		switch pgErr.Code {
		case "23505":
			return domain.Account{}, fmt.Errorf("insert account: %w: %w", domain.ErrDuplicateName, err)
		case "23503":
			if pgErr.ConstraintName == "accounts_user_id_fkey" {
				return domain.Account{}, fmt.Errorf("insert account: %w: %w", domain.ErrUserNotFound, err)
			}

			if pgErr.ConstraintName == "accounts_currency_fkey" {
				return domain.Account{}, fmt.Errorf("insert account: %w: %w", domain.ErrInvalidReference, err)
			}
		case "22P02":
			return domain.Account{}, fmt.Errorf("insert account: %w: %w", domain.ErrInvalidInput, err)
		}
	}

	return domain.Account{}, fmt.Errorf("insert account: %w", err)
}

func (r *Repository) ListAccounts(ctx context.Context, userID uuid.UUID) ([]domain.Account, error) {
	const query = `SELECT id, name, account_type, currency, user_id, opening_balance, created_at FROM accounts WHERE user_id = $1 AND NOT is_archived ORDER BY name`

	rows, err := r.pool.Query(ctx, query, userID)

	if err != nil {
		return nil, fmt.Errorf("query accounts: %w", err)
	}

	defer rows.Close()

	accounts := make([]domain.Account, 0)

	for rows.Next() {
		var account domain.Account
		var openingBalance pgtype.Numeric

		if err := rows.Scan(&account.ID, &account.Name, &account.AccountType, &account.Currency, &account.UserID, &openingBalance, &account.CreatedAt); err != nil {
			return nil, fmt.Errorf("scan account: %w", err)
		}

		money, err := openingBalance.Value()

		if err != nil {
			return nil, fmt.Errorf("numeric to string: %w", err)
		}

		account.OpeningBalance = domain.Money(money.(string))
		accounts = append(accounts, account)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("iterate accounts: %w", err)
	}

	return accounts, nil
}

func (r *Repository) UserExists(ctx context.Context, userID uuid.UUID) (bool, error) {
	const query = `SELECT EXISTS(SELECT 1 FROM users WHERE id = $1)`

	var exists bool

	err := r.pool.QueryRow(ctx, query, userID).Scan(&exists)

	if err != nil {
		return false, fmt.Errorf("user exists: %w", err)
	}

	return exists, nil
}

func (r *Repository) CreateCategory(ctx context.Context, in domain.CreateCategoryInput) (domain.Category, error) {
	const query = `INSERT INTO categories (name, kind, color, icon, parent_id, user_id) VALUES ($1, $2, $3, $4, $5, $6) RETURNING id, created_at`

	category := domain.Category{
		Name:     in.Name,
		Kind:     in.Kind,
		Color:    in.Color,
		Icon:     in.Icon,
		ParentID: in.ParentID,
		UserID:   in.UserID.String(),
	}

	err := r.pool.QueryRow(ctx, query, in.Name, in.Kind, in.Color, in.Icon, in.ParentID, in.UserID).Scan(&category.ID, &category.CreatedAt)

	if err == nil {
		return category, nil
	}

	if pgErr, ok := errors.AsType[*pgconn.PgError](err); ok {
		switch pgErr.Code {
		case "23505":
			return domain.Category{}, fmt.Errorf("insert category: %w: %w", domain.ErrDuplicateName, err)
		case "23503":
			if pgErr.ConstraintName == "categories_user_id_fkey" {
				return domain.Category{}, fmt.Errorf("insert category: %w: %w", domain.ErrUserNotFound, err)
			}
		case "22P02":
			return domain.Category{}, fmt.Errorf("insert category: %w: %w", domain.ErrInvalidInput, err)
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
