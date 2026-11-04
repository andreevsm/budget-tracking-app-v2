package domain

type CreateAccountInput struct {
	UserID      string `json:"user_id"`
	Name        string `json:"name"`
	AccountType string `json:"account_type"`
	Currency    string `json:"currency"`
}
