package domain

type StatBucket string

const (
	BucketDay   StatBucket = "day"
	BucketWeek  StatBucket = "week"
	BucketMonth StatBucket = "month"
	BucketYear  StatBucket = "year"
)

func (b StatBucket) Valid() bool {
	switch b {
	case BucketDay, BucketWeek, BucketMonth, BucketYear:
		return true
	}

	return false
}

type PeriodStat struct {
	Period  Date  `json:"period"`
	Income  Money `json:"income"`
	Expense Money `json:"expense"`
}

type PeriodStatistics struct {
	TotalIncome  Money        `json:"total_income"`
	TotalExpense Money        `json:"total_expense"`
	ByPeriod     []PeriodStat `json:"by_period"`
}
