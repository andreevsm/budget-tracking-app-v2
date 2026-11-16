package domain

import "github.com/google/uuid"

type CreateAccountInput struct {
	UserID      uuid.UUID `json:"user_id"`
	Name        string    `json:"name"`
	AccountType string    `json:"account_type"`
	Currency    string    `json:"currency"`
}
