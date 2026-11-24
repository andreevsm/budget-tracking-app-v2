package domain

import (
	"github.com/google/uuid"
)

type CreateOperationInput struct {
	AccountID   uuid.UUID     `json:"account_id"`
	CategoryID  *uuid.UUID    `json:"category_id"`
	Type        OperationType `json:"type"`
	Amount      Money         `json:"amount"`
	OccurredOn  Date          `json:"occurred_on"`
	Payee       *string       `json:"payee"`
	Description *string       `json:"description"`
}
