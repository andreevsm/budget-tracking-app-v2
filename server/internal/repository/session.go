package repository

import (
	"budget/internal/domain"
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
)

func (r *Repository) CreateSession(ctx context.Context, userID uuid.UUID, tokenHash string, expiresAt time.Time) error {
	const query = `INSERT INTO sessions (user_id, token_hash, expires_at) VALUES ($1, $2, $3)`

	_, err := r.pool.Exec(ctx, query, userID, tokenHash, expiresAt)

	if err != nil {
		return fmt.Errorf("insert session: %w", err)
	}

	return nil
}

func (r *Repository) GetSessionByHash(ctx context.Context, tokenHash string) (domain.Session, error) {
	const query = `SELECT user_id, expires_at, revoked_at FROM sessions WHERE token_hash = $1`

	var session domain.Session

	err := r.pool.QueryRow(ctx, query, tokenHash).Scan(&session.UserID, &session.ExpiresAt, &session.RevokedAt)

	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return domain.Session{}, domain.ErrSessionNotFound
		}

		return domain.Session{}, fmt.Errorf("query session: %w", err)
	}

	return session, nil
}

func (r *Repository) RevokeSession(ctx context.Context, tokenHash string) error {
	const query = `UPDATE sessions SET revoked_at = now() WHERE token_hash = $1 AND revoked_at IS NULL`

	_, err := r.pool.Exec(ctx, query, tokenHash)

	if err != nil {
		return fmt.Errorf("update session: %w", err)
	}

	return nil
}
