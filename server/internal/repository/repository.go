package repository

import (
	"context"
	"fmt"

	"github.com/jackc/pgx/v5/pgxpool"

	"budget/internal/domain"
)

type Repository struct {
	pool *pgxpool.Pool
}

func NewRepository(pool *pgxpool.Pool) *Repository {
	return &Repository{pool: pool}
}

func (r *Repository) List(ctx context.Context) ([]domain.Currency, error) {
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
