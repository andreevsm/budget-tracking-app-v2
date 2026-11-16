package domain

import (
	"errors"
	"time"

	"github.com/google/uuid"
)

type Account struct {
	ID             string    `json:"id"`
	CreatedAt      time.Time `json:"created_at"`
	UserID         uuid.UUID `json:"user_id"`
	Name           string    `json:"name"`
	AccountType    string    `json:"account_type"`
	Currency       string    `json:"currency"`
	OpeningBalance Money     `json:"opening_balance"`
}

var ErrDuplicateName = errors.New("account name already exists")
var ErrInvalidReference = errors.New("invalid reference")
var ErrInvalidInput = errors.New("invalid input")
var ErrUserNotFound = errors.New("user not found")
var ErrAccountNotFound = errors.New("account not found")
