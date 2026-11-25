package domain

import (
	"time"

	"github.com/google/uuid"
)

type Operation struct {
	ID              string        `json:"id"`
	UserID          uuid.UUID     `json:"user_id"`
	AccountID       uuid.UUID     `json:"account_id"`
	CategoryID      *uuid.UUID    `json:"category_id"`
	Type            OperationType `json:"type"`
	Amount          Money         `json:"amount"`
	OccurredOn      Date          `json:"occurred_on"`
	Payee           *string       `json:"payee"`
	Description     *string       `json:"description"`
	TransferGroupID *uuid.UUID    `json:"transfer_group_id,omitempty"`
	CreatedAt       time.Time     `json:"created_at"`
}

type UpdateOperationInput struct {
	Amount      Money      `json:"amount"`
	CategoryID  *uuid.UUID `json:"category_id"`
	OccurredOn  Date       `json:"occurred_on"`
	Payee       *string    `json:"payee"`
	Description *string    `json:"description"`
}
