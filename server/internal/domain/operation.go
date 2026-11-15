package domain

import (
	"time"

	"github.com/google/uuid"
)

type Operation struct {
	ID          string        `json:"id"`
	UserID      uuid.UUID     `json:"user_id"`
	AccountID   uuid.UUID     `json:"account_id"`
	CategoryID  *uuid.UUID    `json:"category_id"`
	Type        OperationType `json:"type"`
	Amount      Money         `json:"amount"`
	OccurredOn  Date          `json:"occurred_on"`
	Payee       *string       `json:"payee"`
	Description *string       `json:"description"`
	CreatedAt   time.Time     `json:"created_at"`
}
