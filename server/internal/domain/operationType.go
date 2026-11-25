package domain

type OperationType string

const (
	Income   OperationType = "income"
	Expense  OperationType = "expense"
	Transfer OperationType = "transfer"
)
