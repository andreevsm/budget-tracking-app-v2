package service

import (
	"budget/internal/domain"
	"budget/internal/repository"
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

var CategoryKinds = []domain.OperationType{
	domain.Income,
	domain.Expense,
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

func (s *Service) ListAccounts(ctx context.Context, userID string) ([]domain.Account, error) {
	userUUID, err := uuid.Parse(userID)

	if err != nil {
		return []domain.Account{}, fmt.Errorf("list accounts: %w: %v", domain.ErrInvalidInput, "user_id is not correct")
	}

	userExists, err := s.repo.UserExists(ctx, userUUID)

	if err != nil {
		return []domain.Account{}, fmt.Errorf("list accounts: %w", err)
	}

	if !userExists {
		return []domain.Account{}, fmt.Errorf("list accounts: %w: %v", domain.ErrUserNotFound, "user doesn't exist")
	}

	accounts, err := s.repo.ListAccounts(ctx, userUUID)

	if err != nil {
		return []domain.Account{}, err
	}

	return accounts, nil
}

func (s *Service) ListOperationsByAccount(ctx context.Context, userID, accountID string) ([]domain.Operation, error) {
	userUUID, err := uuid.Parse(userID)

	if err != nil {
		return []domain.Operation{}, fmt.Errorf("list operations: %w: %v", domain.ErrInvalidInput, "user_id is not correct")
	}

	accountUUID, err := uuid.Parse(accountID)

	if err != nil {
		return []domain.Operation{}, fmt.Errorf("list operations: %w: %v", domain.ErrInvalidInput, "account_id is not correct")
	}

	if accountUUID == uuid.Nil {
		return []domain.Operation{}, fmt.Errorf("list operations: %w: %v", domain.ErrInvalidInput, "account_id is missing")
	}

	accountExists, err := s.repo.AccountBelongsToUser(ctx, accountUUID, userUUID)

	if err != nil {
		return []domain.Operation{}, err
	}

	if !accountExists {
		return []domain.Operation{}, fmt.Errorf("list operations: %w: %v", domain.ErrAccountNotFound, "account doesn't exist")
	}

	operations, err := s.repo.ListOperationsByAccount(ctx, accountUUID)

	if err != nil {
		return []domain.Operation{}, err
	}

	return operations, nil
}

func (s *Service) CreateCategory(ctx context.Context, in domain.CreateCategoryInput) (domain.Category, error) {
	if in.UserID == uuid.Nil {
		return domain.Category{}, fmt.Errorf("create category: %w: %v", domain.ErrInvalidInput, "user_id is empty")
	}

	if in.ParentID != nil {
		return domain.Category{}, fmt.Errorf("create category: %w: %v", domain.ErrInvalidInput, "parent_id must be null")
	}

	name := strings.TrimSpace(in.Name)

	if name == "" {
		return domain.Category{}, fmt.Errorf("create category: %w: %v", domain.ErrInvalidInput, "name should not be empty")
	}

	if utf8.RuneCountInString(name) > 100 {
		return domain.Category{}, fmt.Errorf("create category: %w: %v", domain.ErrInvalidInput, "name length should be at most 100 characters")
	}

	isValidCategoryKind := false

	for _, categoryKind := range CategoryKinds {
		if categoryKind == in.Kind {
			isValidCategoryKind = true
			break
		}
	}

	if !isValidCategoryKind {
		return domain.Category{}, fmt.Errorf("create category: %w: %v", domain.ErrInvalidInput, "kind should have correct type")
	}

	category, err := s.repo.CreateCategory(ctx, domain.CreateCategoryInput{
		Name:     name,
		UserID:   in.UserID,
		Kind:     in.Kind,
		Color:    in.Color,
		Icon:     in.Icon,
		ParentID: in.ParentID,
	})

	if err != nil {
		return domain.Category{}, err
	}

	return category, nil
}

func (s *Service) CreateOperation(ctx context.Context, in domain.CreateOperationInput) (domain.Operation, error) {
	if in.Amount.Valid() {
		if in.Type == domain.Expense && in.Amount.Sign() != -1 {
			return domain.Operation{}, fmt.Errorf("create operation: %w: %v", domain.ErrInvalidInput, "amount with type 'expense' must be < 0")
		}

		if in.Type == domain.Income && in.Amount.Sign() != 1 {
			return domain.Operation{}, fmt.Errorf("create operation: %w: %v", domain.ErrInvalidInput, "amount with type 'income' must be > 0")
		}
	} else {
		return domain.Operation{}, fmt.Errorf("create operation: %w: %v", domain.ErrInvalidInput, "amount must be correct")
	}

	if !in.OccurredOn.Valid() {
		return domain.Operation{}, fmt.Errorf("create operation: %w: %v", domain.ErrInvalidInput, "occurred_on must be correct")
	}

	if in.AccountID == uuid.Nil {
		return domain.Operation{}, fmt.Errorf("create operation: %w: %v", domain.ErrInvalidInput, "account_id is missing")
	}

	accountExists, err := s.repo.AccountBelongsToUser(ctx, in.AccountID, in.UserID)

	if err != nil {
		return domain.Operation{}, err
	}

	if !accountExists {
		return domain.Operation{}, fmt.Errorf("create operation: %w: %v", domain.ErrAccountNotFound, "account doesn't exist")
	}

	operation, err := s.repo.CreateOperation(ctx, domain.CreateOperationInput{
		UserID:      in.UserID,
		AccountID:   in.AccountID,
		CategoryID:  in.CategoryID,
		Type:        in.Type,
		Amount:      in.Amount,
		OccurredOn:  in.OccurredOn,
		Payee:       in.Payee,
		Description: in.Description,
	})

	if err != nil {
		return domain.Operation{}, err
	}

	return operation, nil
}

func (s *Service) ListCategories(ctx context.Context, userID string) ([]domain.Category, error) {
	userUUID, err := uuid.Parse(userID)

	if err != nil {
		return []domain.Category{}, fmt.Errorf("list categories: %w: %v", domain.ErrInvalidInput, "user_id is not correct")
	}

	userExists, err := s.repo.UserExists(ctx, userUUID)

	if err != nil {
		return []domain.Category{}, fmt.Errorf("list categories: %w", err)
	}

	if !userExists {
		return []domain.Category{}, fmt.Errorf("list categories: %w: %v", domain.ErrUserNotFound, "user doesn't exist")
	}

	categories, err := s.repo.ListCategories(ctx, userUUID)

	if err != nil {
		return []domain.Category{}, err
	}

	return categories, nil
}
