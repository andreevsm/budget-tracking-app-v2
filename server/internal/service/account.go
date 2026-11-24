package service

import (
	"budget/internal/domain"
	"context"
	"fmt"
	"strings"
	"unicode/utf8"

	"github.com/google/uuid"
)

type AccountType string

const (
	AccountTypeChecking   AccountType = "checking"
	AccountTypeSavings    AccountType = "savings"
	AccountTypeCreditCard AccountType = "credit_card"
	AccountTypeCash       AccountType = "cash"
	AccountTypeInvestment AccountType = "investment"
	AccountTypeLoan       AccountType = "loan"
	AccountTypeOther      AccountType = "other"
)

var AccountTypes = []AccountType{
	AccountTypeChecking,
	AccountTypeSavings,
	AccountTypeCreditCard,
	AccountTypeCash,
	AccountTypeInvestment,
	AccountTypeLoan,
	AccountTypeOther,
}

func (s *Service) CreateAccount(ctx context.Context, userID uuid.UUID, in domain.CreateAccountInput) (domain.Account, error) {
	name := strings.TrimSpace(in.Name)

	if name == "" {
		return domain.Account{}, fmt.Errorf("create account: %w: %v", domain.ErrInvalidInput, "name should not be empty")
	}

	if utf8.RuneCountInString(name) > 100 {
		return domain.Account{}, fmt.Errorf("create account: %w: %v", domain.ErrInvalidInput, "name length should be at most 100 characters")
	}

	if utf8.RuneCountInString(in.Currency) != 3 {
		return domain.Account{}, fmt.Errorf("create account: %w: %v", domain.ErrInvalidInput, "currency should have 3 chars")
	}

	for _, char := range in.Currency {
		if char < 'A' || char > 'Z' {
			return domain.Account{}, fmt.Errorf("create account: %w: %v", domain.ErrInvalidInput, "currency must contain only uppercase letters")
		}
	}

	isValidAccountType := false

	for _, accountType := range AccountTypes {
		if accountType == AccountType(in.AccountType) {
			isValidAccountType = true
			break
		}
	}

	if !isValidAccountType {
		return domain.Account{}, fmt.Errorf("create account: %w: %v", domain.ErrInvalidInput, "account_type should have correct type")
	}

	account, err := s.repo.CreateAccount(ctx, userID, domain.CreateAccountInput{
		Name:        name,
		AccountType: in.AccountType,
		Currency:    in.Currency,
	})

	if err != nil {
		return domain.Account{}, err
	}

	return account, nil
}

func (s *Service) ListAccounts(ctx context.Context, userID uuid.UUID) ([]domain.Account, error) {
	userExists, err := s.repo.UserExists(ctx, userID)

	if err != nil {
		return []domain.Account{}, fmt.Errorf("list accounts: %w", err)
	}

	if !userExists {
		return []domain.Account{}, fmt.Errorf("list accounts: %w: %v", domain.ErrUserNotFound, "user doesn't exist")
	}

	accounts, err := s.repo.ListAccounts(ctx, userID)

	if err != nil {
		return []domain.Account{}, err
	}

	return accounts, nil
}
