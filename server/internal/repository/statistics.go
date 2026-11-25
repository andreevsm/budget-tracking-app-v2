package repository

import (
	"budget/internal/domain"
	"context"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgtype"
)

var bucketTrunc = map[domain.StatBucket]string{
	domain.BucketDay:   "day",
	domain.BucketWeek:  "week",
	domain.BucketMonth: "month",
	domain.BucketYear:  "year",
}

func (r *Repository) AccountStatistics(ctx context.Context, accountID, userID uuid.UUID, from, to domain.Date, bucket domain.StatBucket) (domain.PeriodStatistics, error) {
	trunc, ok := bucketTrunc[bucket]
	if !ok {
		return domain.PeriodStatistics{}, fmt.Errorf("account statistics: %w: %v", domain.ErrInvalidInput, "unknown bucket")
	}

	const where = `account_id = $1 AND user_id = $2 AND deleted_at IS NULL AND occurred_on BETWEEN $3 AND $4`

	query := fmt.Sprintf(
		`SELECT date_trunc('%s', occurred_on)::date AS period,
			COALESCE(SUM(amount) FILTER (WHERE type = 'income'), 0)::numeric(20, 4) AS income,	
			COALESCE(SUM(-amount) FILTER (WHERE type = 'expense'), 0)::numeric(20, 4) AS expense	
		FROM transactions
		WHERE %s
		GROUP BY 1
		ORDER BY 1`, trunc, where)

	rows, err := r.pool.Query(ctx, query, accountID, userID, from, to)
	if err != nil {
		return domain.PeriodStatistics{}, fmt.Errorf("query period stats: %w", err)
	}

	defer rows.Close()

	byPeriod := make([]domain.PeriodStat, 0)

	for rows.Next() {
		var period time.Time
		var income, expense pgtype.Numeric

		if err := rows.Scan(&period, &income, &expense); err != nil {
			return domain.PeriodStatistics{}, fmt.Errorf("scan period stat: %w", err)
		}

		incomeMoney, err := moneyFromNumeric(income)
		if err != nil {
			return domain.PeriodStatistics{}, err
		}

		expenseMoney, err := moneyFromNumeric(expense)
		if err != nil {
			return domain.PeriodStatistics{}, err
		}

		byPeriod = append(byPeriod, domain.PeriodStat{
			Period:  domain.Date(period.Format("2006-01-02")),
			Income:  incomeMoney,
			Expense: expenseMoney,
		})
	}

	if err := rows.Err(); err != nil {
		return domain.PeriodStatistics{}, fmt.Errorf("iterate period stats: %w", err)
	}

	totalsQuery := fmt.Sprintf(
		`SELECT COALESCE(SUM(amount) FILTER (WHERE type = 'income'), 0)::numeric(20,4),
			COALESCE(SUM(-amount) FILTER (WHERE type = 'expense'), 0)::numeric(20,4)
    FROM transactions
    WHERE %s`, where)

	var totalIncome, totalExpense pgtype.Numeric
	if err := r.pool.QueryRow(ctx, totalsQuery, accountID, userID, from, to).Scan(&totalIncome, &totalExpense); err != nil {
		return domain.PeriodStatistics{}, fmt.Errorf("query totals: %w", err)
	}

	incomeMoney, err := moneyFromNumeric(totalIncome)
	if err != nil {
		return domain.PeriodStatistics{}, err
	}
	expenseMoney, err := moneyFromNumeric(totalExpense)
	if err != nil {
		return domain.PeriodStatistics{}, err
	}

	return domain.PeriodStatistics{
		TotalIncome:  incomeMoney,
		TotalExpense: expenseMoney,
		ByPeriod:     byPeriod,
	}, nil
}

func moneyFromNumeric(n pgtype.Numeric) (domain.Money, error) {
	v, err := n.Value()
	if err != nil {
		return "", fmt.Errorf("numeric to string: %w", err)
	}
	return domain.Money(v.(string)), nil
}
