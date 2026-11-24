package domain

import "github.com/google/uuid"

type CreateTransferInput struct {
	FromAccountID uuid.UUID `json:"from_account_id"`
	ToAccountID   uuid.UUID `json:"to_account_id"`
	Amount        Money     `json:"amount"`
	OccurredOn    Date      `json:"occurred_on"`
	Payee         *string   `json:"payee"`
	Description   *string   `json:"description"`
}
