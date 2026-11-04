package domain

import (
	"errors"
	"time"
)

type Account struct {
	ID          string    `json:"id"`
	CreatedAt   time.Time `json:"created_at"`
	UserID      string    `json:"user_id"`
	Name        string    `json:"name"`
	AccountType string    `json:"account_type"`
	Currency    string    `json:"currency"`
}

var ErrDuplicateName = errors.New("account name already exists")
var ErrInvalidReference = errors.New("invalid reference")
var ErrInvalidInput = errors.New("invalid input")
var ErrUserNotFound = errors.New("user not found")
