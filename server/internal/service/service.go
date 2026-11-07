package service

import (
	"budget/internal/domain"
	"budget/internal/repository"
	"context"
	"fmt"
	"strings"
	"unicode/utf8"
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

type Service struct {
	repo *repository.Repository
}

func NewService(r *repository.Repository) *Service {
	return &Service{repo: r}
}

func (s *Service) CreateAccount(ctx context.Context, in domain.CreateAccountInput) (domain.Account, error) {
	if in.UserID == "" {
		return domain.Account{}, fmt.Errorf("create account: %w: %v", domain.ErrInvalidInput, "user_id is empty")
	}

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

	account, err := s.repo.CreateAccount(ctx, domain.CreateAccountInput{
		Name:        name,
		UserID:      in.UserID,
		AccountType: in.AccountType,
		Currency:    in.Currency,
	})

	if err != nil {
		return domain.Account{}, err
	}

	return account, nil
}
