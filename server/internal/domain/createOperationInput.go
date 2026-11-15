package domain

import (
	"time"

	"github.com/google/uuid"
)

type CreateOperationInput struct {
	UserID      uuid.UUID     `json:"user_id"`
	AccountID   uuid.UUID     `json:"account_id"`
	CategoryID  *uuid.UUID    `json:"category_id"`
	Type        OperationType `json:"type"`
	Amount      Money         `json:"amount"`
	OccurredOn  time.Time     `json:"occurred_on"`
	Payee       *string       `json:"payee"`
	Description *string       `json:"description"`
}
