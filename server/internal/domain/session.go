package domain

import (
	"time"

	"github.com/google/uuid"
)

type Session struct {
	UserID    uuid.UUID  `json:"user_id"`
	ExpiresAt time.Time  `json:"expires_at"`
	RevokedAt *time.Time `json:"revoked_at"`
}
